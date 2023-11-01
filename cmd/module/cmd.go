package main

import (
	"context"

	"github.com/edaniels/golog"

	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/module"

	"github.com/erh/viamgpsd"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}
func realMain() error {

	ctx := context.Background()
	logger := golog.NewDevelopmentLogger("client")

	myMod, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}

	err = myMod.AddModelFromRegistry(ctx, movementsensor.API, viamgpsd.MovementSensorModel)
	if err != nil {
		return err
	}

	err = myMod.Start(ctx)
	defer myMod.Close(ctx)
	if err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}
