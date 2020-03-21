package logging

//LoggerInterface logger 
type LoggerInterface interface {

	// Tracef formats message according to format specifier
	// and writes to log with level = Trace.
	Tracef(format string, params ...interface{})

	// Debugf formats message according to format specifier
	// and writes to log with level = Debug.
	Debugf(format string, params ...interface{})

	// Infof formats message according to format specifier
	// and writes to log with level = Info.
	Infof(format string, params ...interface{})

	// Warnf formats message according to format specifier
	// and writes to log with level = Warn.
	Warnf(format string, params ...interface{}) error

	// Errorf formats message according to format specifier
	// and writes to log with level = Error.
	Errorf(format string, params ...interface{}) error

	

	// Trace formats message using the default formats for its operands
	// and writes to log with level = Trace
	Trace(v ...interface{})

	// Debug formats message using the default formats for its operands
	// and writes to log with level = Debug
	Debug(v ...interface{})

	// Info formats message using the default formats for its operands
	// and writes to log with level = Info
	Info(v ...interface{})

	// Warn formats message using the default formats for its operands
	// and writes to log with level = Warn
	Warn(v ...interface{}) error

	// Error formats message using the default formats for its operands
	// and writes to log with level = Error
	Error(v ...interface{}) error
}


type defaultLogger struct {

}

	// Tracef formats message according to format specifier
	// and writes to log with level = Trace.
func (d *defaultLogger)	Tracef(format string, params ...interface{}){

}

// Debugf formats message according to format specifier
// and writes to log with level = Debug.
func (d *defaultLogger)Debugf(format string, params ...interface{}){
	
}

// Infof formats message according to format specifier
// and writes to log with level = Info.
func (d *defaultLogger)Infof(format string, params ...interface{}){

}

// Warnf formats message according to format specifier
// and writes to log with level = Warn.
func (d *defaultLogger)Warnf(format string, params ...interface{}) error{
	return nil
}

// Errorf formats message according to format specifier
// and writes to log with level = Error.
func (d *defaultLogger)Errorf(format string, params ...interface{}) error{
	return nil
}


// Trace formats message using the default formats for its operands
// and writes to log with level = Trace
func (d *defaultLogger)Trace(v ...interface{}){

}

// Debug formats message using the default formats for its operands
// and writes to log with level = Debug
func (d *defaultLogger)Debug(v ...interface{}){

}

// Info formats message using the default formats for its operands
// and writes to log with level = Info
func (d *defaultLogger)Info(v ...interface{}){

}

// Warn formats message using the default formats for its operands
// and writes to log with level = Warn
func (d *defaultLogger)Warn(v ...interface{}) error{
	return nil
}

// Error formats message using the default formats for its operands
// and writes to log with level = Error
func (d *defaultLogger)Error(v ...interface{}) error{
	return nil
}

var current LoggerInterface 

func init() {
	current = &defaultLogger{}
}
//SetLogger set logger 
func SetLogger( logger LoggerInterface){
	current = logger
}


// Tracef formats message according to format specifier
// and writes to log with level = Trace.
func Tracef(format string, params ...interface{}){
	current.Tracef(format,params...)
}

// Debugf formats message according to format specifier
// and writes to log with level = Debug.
func	Debugf(format string, params ...interface{}){
	current.Debugf(format,params...)
}

// Infof formats message according to format specifier
// and writes to log with level = Info.
func	Infof(format string, params ...interface{}){
	current.Infof(format,params...)
}

// Warnf formats message according to format specifier
// and writes to log with level = Warn.
func	Warnf(format string, params ...interface{}) error{
	return current.Warnf(format,params...)
}

// Errorf formats message according to format specifier
// and writes to log with level = Error.
func	Errorf(format string, params ...interface{}) error{
	return current.Errorf(format,params...)
}



// Trace formats message using the default formats for its operands
// and writes to log with level = Trace
func Trace(v ...interface{}){
	current.Trace(v...)
}

// Debug formats message using the default formats for its operands
// and writes to log with level = Debug
func Debug(v ...interface{}){
	current.Debug(v...)
}

// Info formats message using the default formats for its operands
// and writes to log with level = Info
func Info(v ...interface{}){
	current.Info(v...)
}

// Warn formats message using the default formats for its operands
// and writes to log with level = Warn
func Warn(v ...interface{}) error{
	return current.Warn(v...)
}

// Error formats message using the default formats for its operands
// and writes to log with level = Error
func Error(v ...interface{}) error{
	return current.Error(v...)
}