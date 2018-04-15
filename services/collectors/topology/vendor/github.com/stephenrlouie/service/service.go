// A Service is a long or short runnng go routine to be dispatched by a
// parent routine. This package provides a framework to make managing child
// routines easier. The Service interface must be implemented
// and instances of those services will be added to a ServiceGroup.
// ServiceGroups run all containing Services and will stop them all
// upon receiving an error from a child or ServiceGroup.Kill() is called
package service

// Service is the simple interface containing a Start and Stop function
// that allow the parent routine to operate on the children routines
type Service interface {
	Start() error
	Stop()
}
