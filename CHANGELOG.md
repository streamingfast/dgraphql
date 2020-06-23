# Change log

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.0.1] - 2020-06-23

### Fixed
* Fixed client-js not reconnecting on certain circumstances

### Changed
* Now uses new dauth `authenticator` module
* Now uses a custom ErrorLog on HTTP servers

## [pre-open-sourcing]

### Changed
* Default parameters for `--auth-plugin` and `--dmetering-plugin` changed to `null://`.
* License changed to Apache 2.0
