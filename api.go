package gobe

import (
	. "github.com/rafa-mori/gobe/internal/interfaces"
	. "github.com/rafa-mori/gobe/internal/security/interfaces"

	isc "github.com/rafa-mori/gobe/internal/security/certificates"
	t "github.com/rafa-mori/gobe/internal/types"
	l "github.com/rafa-mori/logz"
)

//func StartGoBE(name string, port string, bind string, logFile string, configFile string, isConfidential bool, logger l.Logger, debug bool) {
//	gl.Log("fatal", "Starting GoBE with name: ", name, " and port: ", port)
//
//	sigChan := make(chan os.Signal, 1)
//	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
//
//	// Initialize the logger
//	if logger == nil {
//		logger = l.GetLogger("GoBE")
//	}
//	gb, gbErr := NewGoBE(name, port, bind, logFile, configFile, isConfidential, logger, debug)
//	if gbErr != nil {
//		gl.Log("fatal", "Failed to create GoBE instance: ", gbErr.Error())
//	}
//
//	if gb == nil {
//		gl.Log("fatal", "Failed to create GoBE instance: ", "GoBE instance is nil")
//	} else {
//		err := gb.Initialize()
//		if err != nil {
//			gl.Log("fatal", "Failed to initialize GoBE: ", err.Error())
//			return
//		}
//		gb.StartGoBE()
//		gl.Log("success", "GoBE started successfully")
//		gl.Log("notice", "GoBE is running on ", gb.GetReference().GetID().String(), " with PID ", strconv.Itoa(os.Getpid()))
//	}
//
//	// Wait for termination signal
//	select {
//	case sig := <-sigChan:
//		gl.Log("info", "Received signal: ", sig.String())
//		//gb.Shutdown()
//		gl.Log("info", "GoBE shutting down")
//	case <-time.After(60 * time.Second):
//		gl.Log("debug", "No signal received, continuing to run")
//		//gb.SyncMetrics()
//	}
//}

type PropertyValBase[T any] interface{ IPropertyValBase[T] }
type Property[T any] interface{ IProperty[T] }

func NewProperty[T any](name string, v *T, withMetrics bool, cb func(any) (bool, error)) IProperty[T] {
	return t.NewProperty(name, v, withMetrics, cb)
}

type Channel[T any] interface{ IChannelCtl[T] }
type ChannelBase[T any] interface{ IChannelBase[T] }

func NewChannel[T any](name string, logger l.Logger) IChannelCtl[T] {
	return t.NewChannelCtl[T](name, logger)
}
func NewChannelCtlWithProperty[T any, P IProperty[T]](name string, buffers *int, property P, withMetrics bool, logger l.Logger) IChannelCtl[T] {
	return t.NewChannelCtlWithProperty[T, P](name, buffers, property, withMetrics, logger)
}
func NewChannelBase[T any](name string, buffers int, logger l.Logger) IChannelBase[T] {
	return t.NewChannelBase[T](name, buffers, logger)
}

type Validation[T any] interface{ IValidation[T] }

func NewValidation[T any](name string, v *T, withMetrics bool, cb func(any) (bool, error)) IValidation[T] {
	return t.NewValidation[T]()
}

type ValidationFunc[T any] interface{ IValidationFunc[T] }

func NewValidationFunc[T any](priority int, f func(value *T, args ...any) IValidationResult) IValidationFunc[T] {
	return t.NewValidationFunc[T](priority, f)
}

type ValidationResult interface{ IValidationResult }

func NewValidationResult(isValid bool, message string, metadata map[string]any, err error) IValidationResult {
	return t.NewValidationResult(isValid, message, metadata, err)
}

type Environment interface{ IEnvironment }

func NewEnvironment(envFile string, isConfidential bool, logger l.Logger) (IEnvironment, error) {
	return t.NewEnvironment(envFile, isConfidential, logger)
}

type Mapper[T any] interface{ IMapper[T] }

func NewMapper[T any](object *T, filePath string) IMapper[T] {
	return t.NewMapper[T](object, filePath)
}

type Mutexes interface{ IMutexes }

func NewMutexes() IMutexes    { return t.NewMutexes() }
func NewMutexesType() Mutexes { return t.NewMutexesType() }

type Reference interface{ IReference }

func NewReference(name string) IReference { return t.NewReference(name) }

type SignalManager[T chan string] interface{ ISignalManager[T] }

func NewSignalManager[T chan string](signalChan T, logger l.Logger) ISignalManager[T] {
	return t.NewSignalManager[T](signalChan, logger)
}

type CertService interface{ ICertService }

func NewCertService(keyPath string, certPath string) ICertService {
	return isc.NewCertService(keyPath, certPath)
}

type CryptoService interface{ ICryptoService }

func NewCryptoService() ICryptoService {
	return nil //t.NewCryptoService()
}
