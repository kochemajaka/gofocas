# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added
- Initial implementation: `Dial`, `Close`, `Ping`
- Readers: `System`, `Status`, `Axes`, `Spindles`, `Feed`, `ContourFeedRate`, `FeedOverride`, `JogOverride`, `ExecutingProgram`, `ProgramSource`, `Alarms`, `Parameters`, `Parameter`, `Diagnosis`
- Series support: 0i, 15, 15i, 16, 16i, 18i, 21, 30i, 31i, 32i
- Automatic reconnect on transient errors
- Stub build mode (`!focas_cgo` tag) for CI without the FANUC SDK
- Six runnable examples
