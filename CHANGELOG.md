# Changelog

All notable changes to the Go Apito SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2024-12-30

### Added

- 🎯 **Type-Safe Operations**: Complete generic typed methods for all operations
  - `GetSingleResourceTyped[T]()` for type-safe single resource retrieval
  - `SearchResourcesTyped[T]()` for type-safe search operations
  - `GetRelationDocumentsTyped[T]()` for type-safe relation queries
  - `CreateNewResourceTyped[T]()` for type-safe resource creation
  - `UpdateResourceTyped[T]()` for type-safe resource updates
- 🚀 **Comprehensive Todo Example**: Complete practical example demonstrating all SDK features
  - Authentication & tenant token generation
  - Resource creation (todos, users, categories)
  - Both typed and untyped search operations
  - Single resource retrieval
  - Resource updates with connections
  - Relation document queries
  - Audit logging
  - Debug functionality
  - Resource cleanup
- 📚 **Enhanced Documentation**: Completely rewritten README with comprehensive examples
  - Quick start guide
  - Complete API reference
  - Type system documentation
  - Plugin integration examples
  - Production deployment guides
  - Performance optimization tips
  - Error handling best practices
- 🔧 **Improved Request Structure**: New `CreateAndUpdateRequest` struct for cleaner API
- 📊 **Version Tracking**: Added `version.go` with `GetVersion()` function

### Changed

- 🔄 **Updated Client Interface**: Enhanced all methods to use the new request structure
- 📖 **Documentation**: Complete rewrite with practical examples and comprehensive coverage
- 🎨 **Example Structure**: Replaced basic example with comprehensive todo application

### Fixed

- 🐛 **Type Conversion**: Improved JSON marshaling/unmarshaling for typed operations
- 🔧 **Error Handling**: Enhanced GraphQL and HTTP error reporting

### Technical Details

- All generic functions follow the pattern: `OperationTyped[T](client, ctx, ...params)`
- Backward compatibility maintained for all existing non-typed methods
- Enhanced context support with tenant ID handling
- Improved connection pooling and performance optimizations

## [1.1.3] - Previous Version

- Previous features and bug fixes

## [1.1.2] - Previous Version

- Previous features and bug fixes

## [1.1.1] - Previous Version

- Previous features and bug fixes

## [1.1.0] - Previous Version

- Previous features and bug fixes

## [1.0.0] - Initial Release

- Initial SDK implementation
- Basic GraphQL communication
- API key authentication
- Core CRUD operations
