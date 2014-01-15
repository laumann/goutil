Go utilities
============

A small collection of utility packages, intended to be used in other code.

To use one of these (for example 'directorywatcher') simply do

	import "github.com/laumann/goutil/directorywatcher"

in your Go program.

Packages
--------

The following lists the packages exported from this repo. All these names should
be prefixed with `github.com/laumann/goutil/`.

 * `directorywatcher` provides a simple-to-use directory watching mechanism,
   which provides events on updates on a watched path.

 * `env` provides the available environment variables in a map.

Feel free to copy the code.
