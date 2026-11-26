use serde_json::{json, Map, Value};
use sidereon_core::astro::bodies::{sun_moon_gcrs, sun_moon_gcrs_at};
use sidereon_core::astro::de::{DeBody, DeEphemeris};
use sidereon_core::astro::events::eclipse::{EarthShadowModel, ShadowFunction};
use sidereon_core::astro::forces::{
    ecom_geometry_with_shadow, AccelerationPartials, DragParameters, EarthRadiationPressure,
    Ecom1Coefficients, Ecom1SolarRadiationPressure, Ecom2D4b1Coefficients,
    Ecom2SolarRadiationPressure, EmpiricalHarmonic, EmpiricalRtnAcceleration, ForceModel,
    IcgemGravityField, J2Gravity, RtnAxis, SolarRadiationPressure, SpaceWeather,
    SpaceWeatherSource,
};
use sidereon_core::astro::frames::transforms::{
    gcrs_to_mean_of_date_matrix, mean_of_date_to_gcrs_matrix,
};
use sidereon_core::astro::frames::{
    EarthOrientationProvider, Iers2010EarthOrientationProvider, Iers2010EopSample,
};
use sidereon_core::astro::propagator::analytic::{
    self, BrouwerLyddanePropagator, CircularElements, EcksteinHechlerPropagator, KeplerianElements,
    MeanConversionOptions, ZonalHarmonics,
};
use sidereon_core::astro::propagator::{
    ForceModelComponents, ForceModelKind, IntegratorKind, IntegratorOptions, PropagationContext,
    StatePropagator,
};
use sidereon_core::astro::state::CartesianState;
use sidereon_core::astro::time::{j2000_seconds, TimeScales};
use std::sync::Arc;

fn values(pins: &mut Map<String, Value>, name: &str, values: impl IntoIterator<Item = f64>) {
    pins.insert(name.into(), json!(values.into_iter().collect::<Vec<_>>()));
}

fn circular(elements: CircularElements) -> [f64; 6] {
    [
        elements.a_km,
        elements.ex,
        elements.ey,
        elements.i_rad,
        elements.raan_rad,
        elements.mean_latitude_argument_rad,
    ]
}

fn keplerian(elements: KeplerianElements) -> [f64; 6] {
    [
        elements.a_km,
        elements.e,
        elements.i_rad,
        elements.raan_rad,
        elements.arg_perigee_rad,
        elements.mean_anomaly_rad,
    ]
}

fn state(value: CartesianState) -> impl Iterator<Item = f64> {
    [value.epoch_tdb_seconds]
        .into_iter()
        .chain(value.position_array())
        .chain(value.velocity_array())
}

fn analytic(pins: &mut Map<String, Value>) {
    let epoch = 646_315_200.25;
    let initial = CartesianState::new(epoch, [7000.0, 20.0, 30.0], [0.0, 7.5, 1.0]);
    let gravity = ZonalHarmonics::earth();
    let eh_options = MeanConversionOptions::ECKSTEIN_HECHLER_DEFAULT;
    let bl_options = MeanConversionOptions::BROUWER_LYDDANE_DEFAULT;
    values(
        pins,
        "CIRCULAR",
        circular(
            CircularElements::from_cartesian(
                initial.position_array(),
                initial.velocity_array(),
                gravity.mu_km3_s2,
            )
            .unwrap(),
        ),
    );
    let eh_mean = analytic::eckstein_hechler_mean_elements(&initial, &gravity, eh_options).unwrap();
    let bl_mean = analytic::brouwer_lyddane_mean_elements(&initial, &gravity, bl_options).unwrap();
    values(
        pins,
        "EH_MEAN",
        circular(eh_mean.elements)
            .into_iter()
            .chain([f64::from(eh_mean.iterations)]),
    );
    values(
        pins,
        "BL_MEAN",
        keplerian(bl_mean.elements)
            .into_iter()
            .chain([f64::from(bl_mean.iterations)]),
    );
    values(
        pins,
        "EH_OSCULATING",
        circular(
            analytic::eckstein_hechler_osculating_elements(&eh_mean.elements, &gravity).unwrap(),
        ),
    );
    values(
        pins,
        "BL_OSCULATING",
        keplerian(
            analytic::brouwer_lyddane_osculating_elements(&bl_mean.elements, &gravity).unwrap(),
        ),
    );

    let eh = EcksteinHechlerPropagator::new(&initial, &gravity, eh_options).unwrap();
    let bl = BrouwerLyddanePropagator::new(&initial, &gravity, 0.0, bl_options).unwrap();
    let epochs = [epoch + 120.0, epoch - 60.0];
    values(
        pins,
        "EH_EPHEMERIS",
        eh.ephemeris(&epochs).unwrap().into_iter().flat_map(state),
    );
    values(
        pins,
        "BL_EPHEMERIS",
        bl.ephemeris(&epochs).unwrap().into_iter().flat_map(state),
    );
    values(
        pins,
        "EH_AT_EPOCH",
        circular(eh.osculating_elements(epochs[0]).unwrap()),
    );
    values(
        pins,
        "BL_AT_EPOCH",
        keplerian(bl.osculating_elements(epochs[0]).unwrap()),
    );
}

fn de(pins: &mut Map<String, Value>) {
    let ephemeris =
        DeEphemeris::from_bytes(include_bytes!("../../../testdata/orbital/lnxp1990.440")).unwrap();
    values(
        pins,
        "DE_EARTH_GM",
        [ephemeris.gm_m3_s2(DeBody::Earth).unwrap()],
    );
    let initial = [7000.0, 20.0, 30.0];
    let velocity = [0.0, 7.5, 1.0];
    let epoch = 646_315_800.25;
    let earth = EarthRadiationPressure::new(1.25, 0.02)
        .unwrap()
        .with_moon_eclipse(false);
    let solar = SolarRadiationPressure::new(1.25, 0.02)
        .unwrap()
        .with_shadow_function(ShadowFunction::Conical)
        .with_earth_shadow_model(EarthShadowModel::Spherical);
    let components = ForceModelComponents::earth_two_body()
        .with_earth_radiation_pressure(earth)
        .with_solar_radiation_pressure(solar);
    let propagator = StatePropagator::new(
        epoch,
        initial,
        velocity,
        ForceModelKind::composite(components),
        IntegratorKind::Dp54,
    );
    let context = PropagationContext::new().with_de_ephemeris(std::sync::Arc::new(ephemeris));
    let states = propagator
        .ephemeris_with_context(&[epoch + 60.0], &context)
        .unwrap();
    values(pins, "DE_PROPAGATION", state(states[0].clone()));
}

fn force_partial_values(pins: &mut Map<String, Value>, name: &str, value: AccelerationPartials) {
    values(
        pins,
        name,
        (0..9)
            .map(|i| value.wrt_position[(i / 3, i % 3)])
            .chain((0..9).map(|i| value.wrt_velocity[(i / 3, i % 3)])),
    );
}

fn additional_forces(pins: &mut Map<String, Value>) {
    let epoch = j2000_seconds(2003, 10, 29, 12, 0, 0.0);
    let initial = CartesianState::new(epoch, [6778.137, 20.0, 30.0], [0.0, 7.5, 1.0]);
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
    )
    .unwrap();
    let table = sidereon_core::astro::space_weather::parse_csv(include_str!(
        "../../../testdata/space_weather/SW-All-20260702-trim.csv"
    ))
    .unwrap()
    .value;
    let forces: [Arc<dyn ForceModel>; 2] = [
        Arc::new(
            EmpiricalRtnAcceleration::new()
                .with_term(RtnAxis::AlongTrack, EmpiricalHarmonic::Constant, 1e-6)
                .unwrap(),
        ),
        Arc::new(
            EmpiricalRtnAcceleration::new()
                .with_term(RtnAxis::CrossTrack, EmpiricalHarmonic::Constant, 3e-7)
                .unwrap(),
        ),
    ];
    let context = PropagationContext::default();
    let mut options = IntegratorOptions::default();
    options.dense_output = true;
    let epochs = [epoch, epoch + 250.0, epoch + 640.0];
    let end = epochs[2];
    for (name, space_weather) in [
        ("FIXED_DRAG", None),
        (
            "SOURCED_DRAG",
            Some(SpaceWeatherSource::Table(Arc::new(table))),
        ),
    ] {
        let mut propagator = StatePropagator::new(
            epoch,
            initial.position_array(),
            initial.velocity_array(),
            ForceModelKind::two_body_j2(),
            IntegratorKind::Dp54,
        )
        .with_options(options)
        .with_drag(drag);
        propagator.space_weather = space_weather;
        let states = propagator
            .ephemeris_with_additional_forces(&epochs, &forces, &context)
            .unwrap();
        let result = propagator
            .propagate_to_with_additional_forces(end, &forces, &context)
            .unwrap();
        values(
            pins,
            name,
            states
                .into_iter()
                .flat_map(state)
                .chain(state(result.final_state))
                .chain(state(result.dense.unwrap().eval(epoch + 320.0).unwrap())),
        );
    }
}

fn forces(pins: &mut Map<String, Value>) {
    let j2 = J2Gravity::default();
    let sample = CartesianState::new(646_315_200.25, [7000.0, 20.0, 30.0], [f64::NAN; 3]);
    let context = PropagationContext::default();
    values(
        pins,
        "J2_ACCELERATION",
        j2.acceleration(&sample, &context)
            .unwrap()
            .as_slice()
            .iter()
            .copied(),
    );
    let distant = CartesianState::new(0.0, [1e80, 2e80, 3e80], [0.0; 3]);
    values(
        pins,
        "J2_DISTANT_ACCELERATION",
        j2.acceleration(&distant, &context).unwrap().iter().copied(),
    );
    force_partial_values(
        pins,
        "J2_PARTIALS",
        j2.acceleration_partials(&[], &sample, &context).unwrap(),
    );
    let state = CartesianState::new(646_315_200.25, [7000.0, 20.0, 30.0], [0.0, 7.5, 1.0]);
    let position = state.position_km;
    let velocity = state.velocity_km_s;
    let sun = [150_000_000.0, 20_000_000.0, -10_000_000.0].into();
    let moon = [-100_000.0, 200_000.0, 300_000.0].into();
    let coefficients = [1e-8, 2e-8, 3e-8, 4e-8, 5e-8, 6e-8, 7e-8, 8e-8, 9e-8];
    let ecom1 =
        Ecom1SolarRadiationPressure::new(Ecom1Coefficients::from_array(coefficients)).unwrap();
    let ecom2 =
        Ecom2SolarRadiationPressure::d4b1(Ecom2D4b1Coefficients::from_array(coefficients)).unwrap();
    let solar = SolarRadiationPressure::new(1.25, 0.02).unwrap();
    let earth = EarthRadiationPressure::new(1.25, 0.02)
        .unwrap()
        .with_moon_eclipse(false);
    values(
        pins,
        "SOLAR",
        solar
            .acceleration_with_sun_moon(position, sun, moon)
            .unwrap()
            .iter()
            .copied(),
    );
    values(
        pins,
        "EARTH",
        earth
            .acceleration_with_sun_moon(&state, sun, moon)
            .unwrap()
            .iter()
            .copied(),
    );
    values(
        pins,
        "ECOM1_SUN_MOON",
        ecom1
            .acceleration_with_sun_moon(position, velocity, sun, moon)
            .unwrap()
            .iter()
            .copied(),
    );
    values(
        pins,
        "ECOM2_SUN_MOON",
        ecom2
            .acceleration_with_sun_moon(position, velocity, sun, moon)
            .unwrap()
            .iter()
            .copied(),
    );
    force_partial_values(
        pins,
        "ECOM1_STATE_SUN_MOON",
        ecom1
            .acceleration_partials_with_sun_moon(position, velocity, sun, moon)
            .unwrap(),
    );
    force_partial_values(
        pins,
        "ECOM2_STATE_SUN_MOON",
        ecom2
            .acceleration_partials_with_sun_moon(position, velocity, sun, moon)
            .unwrap(),
    );
    values(
        pins,
        "ECOM1_PARAMETER_SUN_MOON",
        ecom1
            .parameter_partials_with_sun_moon(position, velocity, sun, moon)
            .unwrap()
            .into_iter()
            .flat_map(<[f64; 3]>::from),
    );
    values(
        pins,
        "ECOM2_PARAMETER_SUN_MOON",
        ecom2
            .parameter_partials_with_sun_moon(position, velocity, sun, moon)
            .unwrap()
            .into_iter()
            .flat_map(<[f64; 3]>::from),
    );

    let ecom1 = ecom1.with_moon_eclipse(false);
    let ecom2 = ecom2.with_moon_eclipse(false);
    values(
        pins,
        "ECOM1_SUN",
        ecom1
            .acceleration_with_sun(position, velocity, sun)
            .unwrap()
            .iter()
            .copied(),
    );
    values(
        pins,
        "ECOM2_SUN",
        ecom2
            .acceleration_with_sun(position, velocity, sun)
            .unwrap()
            .iter()
            .copied(),
    );
    force_partial_values(
        pins,
        "ECOM1_STATE_SUN",
        ecom1
            .acceleration_partials_with_sun(position, velocity, sun)
            .unwrap(),
    );
    force_partial_values(
        pins,
        "ECOM2_STATE_SUN",
        ecom2
            .acceleration_partials_with_sun(position, velocity, sun)
            .unwrap(),
    );
    values(
        pins,
        "ECOM1_PARAMETER_SUN",
        ecom1
            .parameter_partials_with_sun(position, velocity, sun)
            .unwrap()
            .into_iter()
            .flat_map(<[f64; 3]>::from),
    );
    values(
        pins,
        "ECOM2_PARAMETER_SUN",
        ecom2
            .parameter_partials_with_sun(position, velocity, sun)
            .unwrap()
            .into_iter()
            .flat_map(<[f64; 3]>::from),
    );

    let geometry = ecom_geometry_with_shadow(
        position,
        velocity,
        sun,
        Some(moon),
        EarthShadowModel::Spherical,
    )
    .unwrap();
    values(
        pins,
        "ECOM_GEOMETRY",
        geometry
            .dyb
            .d
            .into_iter()
            .chain(geometry.dyb.y)
            .chain(geometry.dyb.b)
            .chain([
                geometry.delta_u_rad.unwrap(),
                geometry.argument_of_latitude_rad.unwrap(),
                geometry.lighting_ratio,
            ]),
    );
}

fn frames_and_gravity(pins: &mut Map<String, Value>) {
    let ts = TimeScales::from_utc(2020, 6, 25, 12, 0, 0.0).unwrap();
    values(
        pins,
        "GCRS_TO_MOD",
        gcrs_to_mean_of_date_matrix(&ts)
            .unwrap()
            .into_iter()
            .flatten(),
    );
    values(
        pins,
        "MOD_TO_GCRS",
        mean_of_date_to_gcrs_matrix(&ts)
            .unwrap()
            .into_iter()
            .flatten(),
    );
    let bodies = sun_moon_gcrs_at(&ts).unwrap();
    values(pins, "BODIES_AT", bodies.sun.into_iter().chain(bodies.moon));
    let bodies = sun_moon_gcrs(0.2).unwrap();
    values(pins, "BODIES", bodies.sun.into_iter().chain(bodies.moon));

    let sample = Iers2010EopSample {
        mjd_utc: 59025,
        xp_arcsec: 0.1,
        yp_arcsec: 0.2,
        ut1_minus_utc_s: -0.2,
        lod_s: 0.001,
        dx_arcsec: 0.0001,
        dy_arcsec: -0.0001,
        pole_rates_rad_s: Some([8e-14, -3e-14]),
    };
    let orientation = Iers2010EarthOrientationProvider::new(vec![
        sample,
        Iers2010EopSample {
            mjd_utc: 59026,
            ..sample
        },
    ])
    .unwrap()
    .orientation_at_tdb_seconds(646_315_800.25)
    .unwrap();
    let pole = orientation.polar_motion();
    values(
        pins,
        "IERS_ANGLES",
        [pole.xp_rad, pole.yp_rad]
            .into_iter()
            .chain(orientation.earth_rotation_vector_itrf_rad_s())
            .chain(orientation.iers2010_quaternion().unwrap()),
    );
    values(
        pins,
        "IERS_MATRICES",
        [
            orientation.gcrf_to_itrf_matrix(),
            orientation.itrf_to_gcrf_matrix(),
            orientation.gcrf_to_itrf_rotation_rate_matrix(),
            orientation.itrf_to_gcrf_rotation_rate_matrix(),
        ]
        .into_iter()
        .flatten()
        .flatten(),
    );

    let field = IcgemGravityField::parse(include_str!(
        "../../../testdata/orbital/synthetic_icgem2.gfc"
    ))
    .unwrap();
    let gravity = field.gravity(2, 2).unwrap();
    values(
        pins,
        "ICGEM_ACCELERATION",
        gravity
            .body_fixed_acceleration_at_km_s2([7000.0, 20.0, 30.0], 86400.0)
            .unwrap(),
    );
}

pub fn contracts() -> Value {
    let mut pins = Map::new();
    analytic(&mut pins);
    de(&mut pins);
    forces(&mut pins);
    additional_forces(&mut pins);
    frames_and_gravity(&mut pins);
    Value::Object(pins)
}
