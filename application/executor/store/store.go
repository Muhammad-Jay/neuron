// Package store implements the installed-executor catalog used by the
// application/executor resolver and installer. An installed executor is an
// immutable directory that has been fully verified and atomically committed.
// The store never performs resolution or downloading; it only persists and
// serves installed artifacts.
//
// Layout:
//
//	<root>/
//	    <owner>/<...segments>/<version>/
//	        executor.json
//	        install.json
//	        <extracted artifact files>
//
// The Store interface consumed by the resolve/install pipeline is declared in
// application/executor (consumer side); this package supplies the concrete
// FilesystemStore and the install.json record helpers.
package store
