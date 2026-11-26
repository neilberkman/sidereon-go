mod attitude;
mod contracts;

use serde_json::json;
use sidereon_core::astro::{
    attitude::gnss::{dyb_axes, nominal_yaw},
    bodies::sun_moon::sun_moon_gcrs_at_tdb_seconds,
    covariance::Covariance6,
    de::{DeBody, DeEphemeris},
    events::eclipse::{shadow_fraction_with_model_and_gradient, EarthShadowModel},
    forces::{
        DragParameters, EarthRadiationPressure, Ecom1Coefficients, Ecom1SolarRadiationPressure,
        EmpiricalHarmonic, EmpiricalRtnAcceleration, ForceModel, IcgemGravityField, RtnAxis,
        SolarRadiationPressure, SpaceWeather, SphericalHarmonicGravity, ThirdBodyGravity,
    },
    frames::iers2010::{Iers2010EarthOrientationProvider, Iers2010EopSample},
    frames::TdbEarthOrientationProvider,
    propagator::{
        analytic::{
            BrouwerLyddanePropagator, CircularElements, EcksteinHechlerPropagator,
            KeplerianElements, ZonalHarmonics,
        },
        decay::{estimate_decay, DecayConfig},
        driver::{propagate_states, PropagationConfig, PropagationForceModel},
        numerical::ForceModelKind,
        PropagationContext,
    },
    state::CartesianState,
    time::model::{Instant, TimeScale},
    time::scales::{tdb_from_tt_j2000_seconds, tdb_minus_tt_seconds, tt_from_tdb_j2000_seconds},
};
use sidereon_core::orbit_determination::{
    fit_inertial_orbit_parameters, fit_sp3_ecef_orbit_parameters, fit_sp3_ecef_precise_orbit,
    InertialPositionSample, OrbitFitOptions, OrbitParameterCovariance, OrbitParameterFitError,
    OrbitParameterFitOptions,
};
use sidereon_core::{
    ephemeris::Sp3,
    orbit_scoring::{
        score_orbit, OrbitScoreOptions, PredictedSp3, SisreWeightSelection, SisreWeights,
    },
};
use std::{error::Error, sync::Arc};

fn exact_float(value: f64) -> serde_json::Value {
    json!({"decimal": value.to_string(), "bits_hex": format!("{:016x}", value.to_bits())})
}

fn main() -> Result<(), Box<dyn Error>> {
    let epoch = 646315200.25;
    let position = [7000.0, 20.0, 30.0];
    let velocity = [0.0, 7.5, 1.0];
    let sun = [150000000.0, 20000000.0, -10000000.0];
    let moon = [-100000.0, 200000.0, 300000.0];
    let state = CartesianState::new(epoch, position, velocity);
    let context = PropagationContext::default();
    let bodies = sun_moon_gcrs_at_tdb_seconds(epoch)?;
    let radiation = SolarRadiationPressure::new(1.25, 0.02)?;
    let earth_radiation = EarthRadiationPressure::new(1.25, 0.02)?;
    let ecom = Ecom1SolarRadiationPressure::new(Ecom1Coefficients {
        d0: 1.0e-7,
        ..Ecom1Coefficients::default()
    })?;
    let srp_partials = radiation.acceleration_partials(&[1.25], &state, &context)?;
    let field = SphericalHarmonicGravity::egm96_truncated(36, 36)?;
    let drag = DragParameters::from_area_mass(
        2.2,
        10.0,
        1000.0,
        SpaceWeather {
            f107: 70.0,
            f107a: 70.0,
            ap: 4.0,
        },
        120.0,
    )?;
    let drag_state = CartesianState::new(epoch, [6778.137, 20.0, 30.0], [0.0, 4.75, 6.0]);
    let yaw = nominal_yaw(position, velocity, sun)?;
    let axes = dyb_axes(position, sun)?;
    let (fraction, gradient) = shadow_fraction_with_model_and_gradient(
        [-7000.0, 6378.0, 0.0],
        [150000000.0, 0.0, 0.0],
        EarthShadowModel::Wgs84Oblate,
    )?;
    let gravity = ZonalHarmonics::earth();
    let eh = EcksteinHechlerPropagator::from_mean_elements(
        epoch,
        CircularElements {
            a_km: 7200.0,
            ex: 0.001,
            ey: 0.002,
            i_rad: 0.5,
            raan_rad: 0.25,
            mean_latitude_argument_rad: 0.125,
        },
        &gravity,
    )?;
    let bl = BrouwerLyddanePropagator::from_mean_elements(
        epoch,
        KeplerianElements {
            a_km: 7200.0,
            e: 0.01,
            i_rad: 0.5,
            raan_rad: 0.25,
            arg_perigee_rad: 0.75,
            mean_anomaly_rad: 0.125,
        },
        &gravity,
        0.0,
    )?;
    let eh = eh.propagate(epoch + 600.0)?;
    let bl = bl.propagate(epoch + 600.0)?;
    let mut config = PropagationConfig::new(0.0, [7000.0, 0.0, 0.0], [0.0, 7.5, 0.0]);
    let two_body = propagate_states(&config, &[0.0, 60.0, 120.0])?[2];
    config.force_model = PropagationForceModel::TwoBodyJ2;
    let j2 = propagate_states(&config, &[0.0, 60.0, 120.0])?[2];
    let eop = Iers2010EopSample {
        mjd_utc: 59025,
        xp_arcsec: 0.1,
        yp_arcsec: 0.2,
        ut1_minus_utc_s: -0.2,
        lod_s: 0.001,
        dx_arcsec: 0.0001,
        dy_arcsec: -0.0001,
        pole_rates_rad_s: Some([0.0; 2]),
    };
    let eop = Iers2010EarthOrientationProvider::new(vec![
        eop,
        Iers2010EopSample {
            mjd_utc: 59026,
            ..eop
        },
    ])?
    .orientation_at_utc_mjd(59025, 43200.0)?;
    let de = DeEphemeris::from_bytes(include_bytes!("../../../testdata/orbital/lnxp1990.440"))?;
    let coverage = de.coverage_tdb_seconds();
    let de_state = de.sun_moon_at_tdb_seconds(coverage.0 + 3600.0)?;
    let icgem = IcgemGravityField::parse(include_str!(
        "../../../testdata/orbital/synthetic_icgem2.gfc"
    ))?;
    let icgem = icgem.gravity(2, 2)?;
    let arc = PropagationConfig::new(0.0, position, velocity);
    let observations: Vec<_> =
        propagate_states(&arc, &[0.0, 60.0, 120.0, 180.0, 240.0, 300.0, 360.0, 420.0])?
            .into_iter()
            .enumerate()
            .map(|(i, observation)| InertialPositionSample {
                epoch: Instant::from_nanos(TimeScale::Tdb, i as i128 * 60_000_000_000),
                position_gcrf_m: (observation.position_km * 1000.0).into(),
            })
            .collect();
    let mut fit_options = OrbitParameterFitOptions::default();
    fit_options.force_model = ForceModelKind::TwoBody {
        mu_km3_s2: arc.gravitational_parameter(),
    };
    let fit = fit_inertial_orbit_parameters(&observations, "G03".parse()?, None, &fit_options)?;
    let mut ill_conditioned_options = fit_options.clone();
    ill_conditioned_options.max_condition_number = 1.5;
    let ill_conditioned = fit_inertial_orbit_parameters(
        &observations,
        "G03".parse()?,
        None,
        &ill_conditioned_options,
    )
    .expect_err("condition number exceeds the limit");
    let OrbitParameterFitError::IllConditionedNormalEquations {
        satellite,
        condition_number,
        limit,
        weakest,
    } = ill_conditioned
    else {
        return Err(ill_conditioned.into());
    };
    let ill_conditioned = json!({
        "satellite": satellite.to_string(),
        "condition_number": exact_float(condition_number),
        "limit": exact_float(limit),
        "weakest": weakest.into_iter().map(|(name, value)| json!([name, exact_float(value)])).collect::<Vec<_>>(),
    });
    let displaced_seed = CartesianState::new(
        fit.initial_state.epoch_tdb_seconds,
        [position[0] + 0.01, position[1], position[2]],
        velocity,
    );
    let mut non_convergent_options = fit_options.clone();
    non_convergent_options.max_iterations = 1;
    non_convergent_options.convergence_threshold = 1e-12;
    let non_convergent = fit_inertial_orbit_parameters(
        &observations,
        "G03".parse()?,
        Some(displaced_seed),
        &non_convergent_options,
    )
    .expect_err("one iteration cannot converge from the displaced seed");
    let OrbitParameterFitError::DidNotConverge {
        satellite,
        iterations,
        last_step_sigmas,
    } = non_convergent
    else {
        return Err(non_convergent.into());
    };
    let non_convergent = json!({
        "satellite": satellite.to_string(),
        "iterations": iterations,
        "last_step_sigmas": exact_float(last_step_sigmas),
    });
    let converted = KeplerianElements::from_cartesian([4.0, 0.0, 0.0], [0.0, 1.0, 0.0], 4.0)?;
    let sp3 = Sp3::parse(include_bytes!("../../../testdata/trimmed.sp3"))?;
    let cutoff = Instant::from_nanos(TimeScale::Gpst, 646_266_600_000_000_000);
    let prediction = PredictedSp3::with_cutoff(&sp3, cutoff)?;
    let mut score_options = OrbitScoreOptions::new(vec![0.0, 1800.0, 7200.0]);
    score_options.sisre_weights = SisreWeightSelection::Fixed(SisreWeights::new(1.0, 0.02)?);
    let score = score_orbit(&prediction, &sp3, &["G08".parse()?], &score_options)?;
    let sp3_arc = Sp3::parse(include_bytes!("../../../testdata/orbitfit-5epoch.sp3"))?;
    let orientation = TdbEarthOrientationProvider::new();
    let g08 = "G08".parse()?;
    let ecef_fit =
        fit_sp3_ecef_precise_orbit(&sp3_arc, g08, &orientation, &OrbitFitOptions::default())?;
    let ecef_fit = &ecef_fit.fits[&g08].initial_state;
    let ecef_parameters =
        fit_sp3_ecef_orbit_parameters(&sp3_arc, g08, &orientation, None, &fit_options)?;
    let mut force_options = fit_options.clone();
    let parameter = "empirical.along.constant";
    let force = EmpiricalRtnAcceleration::with_zero_terms(&[(
        RtnAxis::AlongTrack,
        EmpiricalHarmonic::Constant,
    )])
    .with_term(RtnAxis::AlongTrack, EmpiricalHarmonic::Constant, 1e-8)?;
    force_options.additional_forces = vec![Arc::new(force)];
    force_options.estimated_parameters = vec![parameter.to_string()];
    let force_fit =
        fit_inertial_orbit_parameters(&observations, "G03".parse()?, None, &force_options)?;
    let fitted_parameter = force_fit
        .force_parameter(parameter)
        .expect("selected parameter");
    let variances = Covariance6::from_diagonal([1.0, 2.0, 3.0, 0.01, 0.02, 0.03])
        .expect("positive covariance variances");
    let propagator = config.to_propagator();
    let (state_with_covariance, covariance) =
        propagator.propagate_state_with_covariance(variances, 120.0)?;
    let stm = propagator.state_transition_matrix_to(120.0)?;
    let descending_state = CartesianState::new(epoch, [6468.137, 0.0, 0.0], [0.0, 7.8, 0.0]);
    let decay = estimate_decay(
        descending_state,
        &DecayConfig::new(drag).with_reentry_altitude_km(125.0),
    )?;
    let OrbitParameterCovariance::Estimated {
        variance_factor,
        degrees_of_freedom,
        ref matrix,
        ..
    } = fit.covariance
    else {
        return Err("fit covariance must be estimated".into());
    };
    let (attitude, attitude_final) = attitude::attitude();
    let output = json!({
        "contracts": contracts::contracts(),
        "attitude": attitude,
        "attitude_final": attitude_final,
        "reentry": { "time": decay.time_to_decay_s, "altitude": decay.reentry_altitude_km },
        "fitted_force": {"Name": fitted_parameter.name, "InitialValue": fitted_parameter.initial_value,
            "Value": fitted_parameter.value},
        "fitted_force_sigma": force_fit.sigma(parameter).expect("force sigma"),
        "sp3_fit": [ecef_fit.position_km.as_slice(), ecef_fit.velocity_km_s.as_slice()],
        "sp3_fit_epoch": ecef_fit.epoch_tdb_seconds,
        "sp3_parameter_fit": [ecef_parameters.initial_state.position_km.as_slice(), ecef_parameters.initial_state.velocity_km_s.as_slice()],
        "sp3_parameter_fit_epoch": ecef_parameters.initial_state.epoch_tdb_seconds,
        "covariance": covariance.into_matrix(),
        "covariance_state": [state_with_covariance.position_km.as_slice(), state_with_covariance.velocity_km_s.as_slice()],
        "state_transition": stm,
        "score_epochs": score.coverage.epochs,
        "score_scored": score.coverage.scored,
        "score_bins": score.per_satellite[0].bins.iter().map(|bin| bin.count).collect::<Vec<_>>(),
        "converted_mean_anomaly": converted.mean_anomaly_rad,
        "de": {
            "version": de.version(),
            "record_size": de.record_size(),
            "start": coverage.0,
            "end": coverage.1,
            "constants": de.constants(),
            "astronomical_unit_m": de.astronomical_unit_m(),
            "earth_moon_mass_ratio": de.earth_moon_mass_ratio(),
            "gm_sun": de.gm_m3_s2(DeBody::Sun),
            "sun_position": de_state.sun.position_m,
            "sun_velocity": de_state.sun.velocity_m_s,
            "moon_position": de_state.moon.position_m,
            "moon_velocity": de_state.moon.velocity_m_s
        },
        "icgem_potential": icgem.body_fixed_potential_at_km2_s2(position.into(), 86400.0)?,
        "iers_quaternion": eop.iers2010_quaternion(),
        "fit": {
            "covariance": {"Estimated": true, "Dimension": matrix.nrows(), "VarianceFactor": variance_factor, "DegreesOfFreedom": degrees_of_freedom},
            "observations": observations.iter().map(|row| row.position_gcrf_m).collect::<Vec<_>>(),
            "initial": [fit.initial_state.position_km.as_slice(), fit.initial_state.velocity_km_s.as_slice()],
            "initial_epoch": fit.initial_state.epoch_tdb_seconds,
            "residual_rms_m": fit.residual_stats.rms_3d_m,
            "seed_rms_m": fit.seed_rms_3d_m,
            "condition_number": fit.condition_number,
            "iterations": fit.iterations,
            "sigma_x_km": fit.sigma("x_km"),
            "correlation_xy": fit.correlation("x_km", "y_km"),
            "ill_conditioned": ill_conditioned,
            "non_convergent": non_convergent,
            "parameter_names": fit.parameter_names,
            "residuals": fit.residuals.iter().map(|row| [row.epoch_tdb_j2000_s, row.elapsed_tt_s, row.radial_m, row.along_m, row.cross_m]).collect::<Vec<_>>(),
        },
        "time": [tdb_minus_tt_seconds(epoch), tdb_from_tt_j2000_seconds(epoch), tt_from_tdb_j2000_seconds(epoch)],
        "sun": bodies.sun,
        "moon": bodies.moon,
        "solar_radiation": radiation.acceleration(&state, &context)?.as_slice(),
        "solar_radiation_partials": std::array::from_fn::<_, 3, _>(|i| std::array::from_fn::<_, 3, _>(|j| srp_partials.wrt_position[(i,j)])),
        "earth_radiation": earth_radiation.acceleration(&state, &context)?.as_slice(),
        "third_body": ThirdBodyGravity::default().acceleration(&state, &context)?.as_slice(),
        "ecom1": ecom.acceleration_with_sun_moon(position.into(), velocity.into(), sun.into(), moon.into())?.as_slice(),
        "gravity_acceleration": field.body_fixed_acceleration_km_s2(position.into())?.as_slice(),
        "gravity_potential": field.body_fixed_potential_km2_s2(position.into())?,
        "drag": drag.to_force().acceleration(&drag_state, &context)?.as_slice(),
        "nominal_yaw": { "YawRad": yaw.yaw_rad, "BetaRad": yaw.beta_rad, "BodyX": yaw.body_x,
            "BodyY": yaw.body_y, "BodyZ": yaw.body_z },
        "dyb": { "D": axes.d, "Y": axes.y, "B": axes.b },
        "shadow_fraction": fraction,
        "shadow_gradient": gradient,
        "eckstein_hechler": [eh.position_km.as_slice(), eh.velocity_km_s.as_slice()],
        "brouwer_lyddane": [bl.position_km.as_slice(), bl.velocity_km_s.as_slice()],
        "two_body": [two_body.position_km.as_slice(), two_body.velocity_km_s.as_slice()],
        "two_body_j2": [j2.position_km.as_slice(), j2.velocity_km_s.as_slice()]
    });
    println!("{}", serde_json::to_string_pretty(&output)?);
    Ok(())
}
