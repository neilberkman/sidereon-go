package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func sourceFixture() ([]SourceSensor, []float64) {
	sensors := []SourceSensor{
		{Dimension: 3, PositionM: [3]float64{0, 0, 0}},
		{Dimension: 3, PositionM: [3]float64{1000, 0, 0}},
		{Dimension: 3, PositionM: [3]float64{0, 1000, 0}},
		{Dimension: 3, PositionM: [3]float64{0, 0, 1000}},
	}
	position := [3]float64{200, 300, 400}
	arrivals := make([]float64, len(sensors))
	for i, sensor := range sensors {
		dx := sensor.PositionM[0] - position[0]
		dy := sensor.PositionM[1] - position[1]
		dz := sensor.PositionM[2] - position[2]
		arrivals[i] = math.Sqrt(dx*dx+dy*dy+dz*dz) / 299792458
	}
	return sensors, arrivals
}

func TestSourceQualityFixture(t *testing.T) {
	sensors, arrivals := sourceFixture()
	guess, err := ClosedFormInitialGuess(sensors, arrivals, 299792458, SourceSolveTOA, 0)
	if err != nil {
		t.Fatal(err)
	}
	if guess.Dimension != 3 || math.Abs(guess.PositionM[0]-200) > 1e-9 || math.Abs(guess.PositionM[1]-300) > 1e-9 || math.Abs(guess.PositionM[2]-400) > 1e-9 || !guess.HasOriginTime || math.Abs(guess.ResidualRMSS) > 1e-15 {
		t.Fatalf("unexpected closed-form guess: %+v", guess)
	}
	guess2, err := ChanHOInitialGuess(sensors, arrivals, 299792458, SourceSolveTOA, 0)
	if err != nil {
		t.Fatal(err)
	}
	if guess2 != guess {
		t.Fatalf("Chan-Ho and closed-form seeds differ: %+v vs %+v", guess2, guess)
	}
	dop, err := SourceDOP(sensors, []float64{200, 300, 400}, 299792458)
	if err != nil {
		t.Fatal(err)
	}
	wantDOP := DOP{GDOP: 4.891083916093432e8, PDOP: 4.891083916093432e8, HDOP: 4.1797252356646854e8, VDOP: 2.540196612196711e8, TDOP: 0.5482279079112733}
	if math.Abs(dop.GDOP-wantDOP.GDOP) > 1e-6 || math.Abs(dop.PDOP-wantDOP.PDOP) > 1e-6 || math.Abs(dop.HDOP-wantDOP.HDOP) > 1e-6 || math.Abs(dop.VDOP-wantDOP.VDOP) > 1e-6 || math.Abs(dop.TDOP-wantDOP.TDOP) > 1e-15 {
		t.Fatalf("unexpected DOP: %+v", dop)
	}
	crlb, err := SourceCRLB(sensors, []float64{200, 300, 400}, 299792458, 1e-9)
	if err != nil {
		t.Fatal(err)
	}
	if crlb.Covariance.Dimension != 3 || crlb.Covariance.StateDimension != 4 || !crlb.Covariance.HasOriginTimeS2 || math.Abs(crlb.Covariance.TimingSigmaS-1e-9) > 1e-20 {
		t.Fatalf("unexpected CRLB: %+v", crlb)
	}
	opts, err := SourceLocateOptionsInit()
	if err != nil {
		t.Fatal(err)
	}
	if opts.Mode != SourceSolveTOA || opts.Loss != SourceLossLinear || opts.TimingSigmaS != 1 || opts.FScaleS != 1 {
		t.Fatalf("unexpected source defaults: %+v", opts)
	}
	sol, err := LocateSourceWith(sensors, arrivals, 299792458, &opts, true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sol.Close() }()
	summary, err := sol.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.Dimension != 3 || summary.ResidualCount != 4 || summary.InfluenceCount != 4 || math.Abs(summary.PositionM[0]-200) > 1e-9 || math.Abs(summary.PositionM[1]-300) > 1e-9 || math.Abs(summary.PositionM[2]-400) > 1e-9 || !summary.HasCovariance {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	residuals, err := sol.Residuals()
	if err != nil {
		t.Fatal(err)
	}
	if len(residuals) != 4 || math.Abs(residuals[0].ResidualS) > 1e-15 {
		t.Fatalf("unexpected residuals: %+v", residuals)
	}
	residuals[0].ResidualS = 123
	again, err := sol.Residuals()
	if err != nil || math.Abs(again[0].ResidualS) > 1e-15 {
		t.Fatalf("residual output was not detached: %+v, %v", again, err)
	}
	influences, err := sol.Influences()
	if err != nil {
		t.Fatal(err)
	}
	if len(influences) != 4 || influences[0].SensorIndex != 0 || influences[0].LossWeight != 1 {
		t.Fatalf("unexpected influences: %+v", influences)
	}
	cov, present, err := sol.Covariance()
	if err != nil {
		t.Fatal(err)
	}
	if !present || cov.Dimension != 3 || cov.StateDimension != 4 {
		t.Fatalf("unexpected covariance: present=%v cov=%+v", present, cov)
	}
	var cn BroadcastCNAV
	cn.Present = true
	cn.URAEDIndex = 0
	cn.URANED0Index = 0
	cn.URANED1Index = 0
	cn.URANED2Index = 0
	cn.Top = GNSSWeekTow{Week: 2000, TOWSeconds: 0}
	ura, uraPresent, err := CNAVURANEDM(cn, 2000, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !uraPresent || ura != 2 {
		t.Fatalf("unexpected NED URA: %v present=%v", ura, uraPresent)
	}
	nominal, nominalPresent, err := CNAVURANominalM(0)
	if err != nil {
		t.Fatal(err)
	}
	if !nominalPresent || nominal != 2 {
		t.Fatalf("unexpected nominal URA: %v present=%v", nominal, nominalPresent)
	}
	po, err := PseudorangeVarianceOptionsInit()
	if err != nil {
		t.Fatal(err)
	}
	variance, err := PseudorangeVariance(90, po)
	if err != nil {
		t.Fatal(err)
	}
	if po.AM != 0.3 || po.BM != 0.3 || po.Model != PseudorangeVarianceElevation || math.Abs(variance-0.18) > 1e-15 {
		t.Fatalf("unexpected pseudorange fixture: options=%+v variance=%v", po, variance)
	}
	vo, err := SolutionValidationOptionsInit()
	if err != nil {
		t.Fatal(err)
	}
	if vo.HasMaxPDOP || vo.MinPlausibleRadiusM != 6.344752e6 || vo.MaxPlausibleRadiusM != 8.378137e6 || vo.MaxConvergedResidualRMSM != 10000 {
		t.Fatalf("unexpected validation defaults: %+v", vo)
	}
}

func TestSourceQualityBoundariesAndCloseRace(t *testing.T) {
	sensors, arrivals := sourceFixture()
	if _, err := ClosedFormInitialGuess(sensors[:3], arrivals[:3], 299792458, SourceSolveTOA, 0); err == nil {
		t.Fatal("expected insufficient-geometry error")
	}
	bad := append([]SourceSensor(nil), sensors...)
	bad[0].Dimension = 4
	if _, err := SourceDOP(bad, []float64{0, 0, 0}, 299792458); err == nil {
		t.Fatal("expected invalid sensor dimension error")
	}
	if _, err := SourceDOP(sensors, []float64{0}, 299792458); err == nil {
		t.Fatal("expected invalid source dimension error")
	}
	if _, err := PseudorangeVariance(90, PseudorangeVarianceOptions{Model: PseudorangeVarianceModel(99)}); err == nil {
		t.Fatal("expected invalid pseudorange model error")
	}
	invalid := SourceLocateOptions{Mode: SourceSolveMode(99), Loss: SourceLossLinear}
	if _, err := LocateSourceWith(sensors, arrivals, 299792458, &invalid, false); err == nil {
		t.Fatal("expected invalid source mode error")
	}
	solution, err := LocateSource(sensors, arrivals, 299792458, nil)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				_, _ = solution.Summary()
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			_ = solution.Close()
		}
	}()
	wg.Wait()
	if _, err := solution.Summary(); !errors.Is(err, ErrClosed) {
		t.Fatalf("summary after close: %v", err)
	}
}

func TestReceiverValidationUsesCFixture(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })
	options, err := SolutionValidationOptionsInit()
	if err != nil {
		t.Fatal(err)
	}
	config := usedSPPConfig()
	config.Validation = &options
	solution, err := SolveSPP(sp3, config)
	if err != nil {
		t.Fatal(err)
	}
	if solution.PositionM == [3]float64{} || solution.UsedSatelliteCount == 0 {
		t.Fatalf("unexpected validated fixture solution: %+v", solution)
	}
	options.MinPlausibleRadiusM = 1e9
	config.Validation = &options
	if _, err := SolveSPP(sp3, config); err == nil {
		t.Fatal("receiver validation unexpectedly accepted impossible radius")
	}
}
