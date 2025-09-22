# Project Plan: Exposing Internal Packages

This document outlines the plan for refactoring the project to expose internal packages as public APIs.

## I. Refactor the Core Client

- [ ] Move the `Client` struct and its methods from `internal/api/client` to `pkg/garth/client`.
- [ ] Update all internal import paths to reflect the new location of the client.
- [ ] Review the public API of the client and ensure that only essential functions are exported.

## II. Expose Data Models

- [ ] Move all data structures from `internal/models/types` to `pkg/garth/types`.
- [ ] Update all internal import paths to reflect the new location of the data types.
- [ ] Review the naming of the data types and ensure that they are consistent and clear.

## III. Make Authentication Logic Public

- [ ] Move the authentication logic from `internal/auth` to `pkg/garth/auth`.
- [ ] Update all internal import paths to reflect the new location of the authentication logic.
- [ ] Create a clean and well-documented public API for the authentication functions.