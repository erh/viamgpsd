package viamgpsd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edaniels/golog"
	"github.com/golang/geo/r3"
	geo "github.com/kellydunn/golang-geo"
	"github.com/stratoberry/go-gpsd"

	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/data"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
)

var MovementSensorModel = resource.ModelNamespace("erh").WithFamily("viamgpsd").WithModel("gps")

func init() {
	resource.RegisterComponent(
		movementsensor.API,
		MovementSensorModel,
		resource.Registration[movementsensor.MovementSensor, resource.NoNativeConfig]{
			Constructor: newMovementSensor,
		})
}

func newMovementSensor(ctx context.Context, deps resource.Dependencies, config resource.Config, logger golog.Logger) (movementsensor.MovementSensor, error) {

	s := &mySensor{
		name:   config.ResourceName(),
		logger: logger,
	}

	s.start(ctx)

	return s, nil
}

type mySensor struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger golog.Logger
	prop   movementsensor.Properties

	session *gpsd.Session

	mu sync.Mutex

	lastData time.Time

	tpv gpsd.TPVReport
}

func (s *mySensor) start(ctx context.Context) error {

	var err error

	if s.session, err = gpsd.Dial(gpsd.DefaultAddress); err != nil {
		return fmt.Errorf("Failed to connect to GPSD: %s", err)
	}

	s.session.AddFilter("TPV", func(r interface{}) {
		tpv := r.(*gpsd.TPVReport)

		s.mu.Lock()
		defer s.mu.Unlock()

		s.prop.PositionSupported = true
		s.prop.CompassHeadingSupported = true
		s.prop.LinearVelocitySupported = true

		s.lastData = time.Now()
		s.tpv = *tpv
	})

	// ATT - will give imu data
	// SKY for accuracy

	s.session.Watch()
	return nil
}

func (s *mySensor) Name() resource.Name {
	return s.name
}

func (s *mySensor) Position(ctx context.Context, extra map[string]interface{}) (*geo.Point, float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return geo.NewPoint(s.tpv.Lat, s.tpv.Lon), s.tpv.Alt, s.tooOld(extra)
}

func (s *mySensor) LinearVelocity(ctx context.Context, extra map[string]interface{}) (r3.Vector, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return r3.Vector{0, s.tpv.Speed, 0}, s.tooOld(extra)
}

func (s *mySensor) AngularVelocity(ctx context.Context, extra map[string]interface{}) (spatialmath.AngularVelocity, error) {
	return spatialmath.AngularVelocity{}, nil
}

func (s *mySensor) LinearAcceleration(ctx context.Context, extra map[string]interface{}) (r3.Vector, error) {
	return r3.Vector{}, nil
}

func (s *mySensor) CompassHeading(ctx context.Context, extra map[string]interface{}) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tpv.Track, s.tooOld(extra)
}

func (s *mySensor) Orientation(ctx context.Context, extra map[string]interface{}) (spatialmath.Orientation, error) {
	return &spatialmath.EulerAngles{}, nil
}

func (s *mySensor) Properties(ctx context.Context, extra map[string]interface{}) (*movementsensor.Properties, error) {
	return &s.prop, nil
}

func (s *mySensor) Accuracy(ctx context.Context, extra map[string]interface{}) (map[string]float32, error) {
	return nil, nil
}

func (s *mySensor) tooOld(extra map[string]interface{}) error {
	if time.Since(s.lastData) < time.Minute {
		return nil
	}

	if extra != nil && extra[data.FromDMString] == true {
		// we're from data capture
		// since data is too old, just don't store anything or log
		return data.ErrNoCaptureToStore
	}

	return fmt.Errorf("lastUpdate update too old: %v (%v)", s.lastData, extra)
}

func (s *mySensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (s *mySensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	return movementsensor.Readings(ctx, s, extra)
}

func (s *mySensor) Close(ctx context.Context) error {
	return s.session.Close()
}
