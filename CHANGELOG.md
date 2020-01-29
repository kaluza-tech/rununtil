# Change Log

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.2] - 2020-01-29

### Fixed

- Fix race condition in CancelAll

## [0.2.1] - 2020-01-28

### Changed

- Migrate to kaluza-tech

## [0.2.0] - 2019-06-12

### Fixed

- There was a problem with sending a kill signal to a nonblocking main function. This has been fixed by providing the CancelAll method.

### Changed

- Deprecate: functions KillSignal, Signals, and Killed
- Replace them with: AwaitKillSignal, AwaitKillSignals and CancelAll

## [0.1.0] - 2019-06-12

### Added

- Implemented the rununtil.Killed method, which allows you to test a function that uses rununtil.KillSignal

### Changed

- There are now RunnerFunc and ShutdownFunc function types to clarify the usage of this library (backwards incompatible change)

## [0.0.1] - 2019-06-05

### Added

- initial commit of rununtil library

