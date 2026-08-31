# Copilot Instructions

## Folder Structure
```
/
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── CONTRIBUTING.md
├── Dockerfile
├── GOLANG_VERSION
├── LICENSE
├── main.go
├── Makefile
├── omnitruck.yml.example
├── README.md
├── sonar-project.properties
├── VERSION
├── .expeditor/
│   └── artifact_upload.sh
├── .github/
│   ├── CODEOWNERS
│   ├── ISSUE_TEMPLATE/
│   └── workflows/
├── clients/
│   ├── iLicense.go
│   ├── license.go
│   ├── license.mock.go
│   ├── request.go
│   └── omnitruck/
│       ├── contains_validator_test.go
│       ├── contains_validator.go
│       ├── dynamo_services_mock.go
│       ├── dynamo_services_test.go
│       ├── dynamo_services.go
│       ├── eol_version_validator.go
│       ├── filters_test.go
│       ├── filters.go
│       ├── idynamo_services.go
│       ├── iplatform_services.go
│       ├── omnitruck_mock.go
│       ├── omnitruck_test.go
│       ├── omnitruck.go
│       ├── os_version_validator.go
│       ├── platform_services_mock.go
│       ├── platform_services_test.go
│       ├── platform_services.go
│       ├── product_test.go
│       ├── product.go
│       ├── validators_mock.go
│       ├── validators_test.go
│       ├── validators.go
│       ├── aws/
│       └── replicated/
├── cmd/
│   ├── root.go
│   └── start.go
├── config/
│   └── config.go
├── constants/
│   └── constants.go
├── dboperations/
│   ├── dboperations_test.go
│   ├── dboperations.go
│   └── mockdboperations.go
├── docs/
│   ├── build.md
│   ├── deploy.md
│   ├── DownloadAPI-Commercial.md
│   ├── DownloadAPI-NonCommercial.md
│   ├── OmnitruckApi_docs.go
│   ├── OmnitruckApi_openapi3.json
│   ├── OmnitruckApi_swagger.json
│   └── OmnitruckApi_swagger.yaml
├── httpserver/
│   ├── routes.go
│   └── server.go
├── internal/
│   ├── api/
│   │   └── handler/
│   │       ├── handler.go
│   │       └── handler_test.go
│   ├── helper/
│   │   ├── helpers_test.go
│   │   └── helpers.go
│   ├── services/
│   │   ├── services_test.go
│   │   └── services.go
│   └── strategy/
│       ├── automate_strategy_test.go
│       ├── automate_strategy.go
│       ├── default_product_strategy_test.go
│       ├── default_product_strategy.go
│       ├── infra_strategy_test.go
│       ├── infra_strategy.go
│       ├── mode_strategy_test.go
│       ├── mode_strategy.go
│       ├── platform_service_strategy_test.go
│       ├── platform_service_strategy.go
│       ├── product_strategy_test.go
│       └── product_strategy.go
├── logger/
│   ├── logger_test.go
│   └── logger.go
├── middleware/
│   ├── db/
│   └── license/
│       └── license.go
├── models/
│   └── models.go
├── scripts/
│   ├── install_golang.sh
│   ├── lamdaFunction.py
│   ├── load-related-products.go
│   ├── push_data_to_database.py
│   └── s3_to_dynamoDB.py
├── static/
│   └── index.html
├── templates/
│   ├── install.ps1.tmpl
│   └── install.sh.tmpl
├── terraform/
│   ├── acceptance/
│   ├── cdn/
│   ├── modules/
│   └── prod/
├── tools/
│   └── go_coverage_report.sh
├── utils/
│   ├── errors.go
│   ├── utils.go
│   ├── awsutils/
│   └── template/
│       └── templateRenderer.go
└── views/
    ├── docs.html
    └── layouts/
```

> **Critical:** Do not modify any `*.codegen.go` files, if codegen files are present.

## JIRA Integration
- When a Jira ID is provided, use the atlassian-mcp-server to fetch the JIRA issue details, read the story, and implement the task.
- Extract requirements, acceptance criteria, and technical specifications from the JIRA issue.
- Use the JIRA story details to guide implementation decisions and ensure all requirements are met.

## Unit Testing Requirements
- Always create comprehensive unit test cases for your implementation.
- Ensure test coverage is above 80% for all new code.
- Use the `testify` framework for assertions and mocks as established in the codebase.
- Mock all external dependencies (AWS services, upstream APIs, database operations).
- Follow existing test patterns found in files like `*_test.go`.
- Run tests using: `go test -race -vet=off ./...`

## Pull Request Workflow
When prompted to create a PR for changes:
1. Use GitHub CLI to create a branch named after the JIRA ID (e.g., `JIRA-123`)
2. Push all changes to the branch
3. Create a PR with a detailed description using HTML tags for formatting
4. All tasks are performed on the local repository
5. **Do not reference `.profile` in GitHub auth steps**

## Prompt-Based Workflow
- All tasks performed should be prompt-based and interactive
- After each step, provide a clear summary of what was accomplished
- Prompt the user with the next step and list remaining steps
- Always ask if the user wants to continue with the next step before proceeding

## Task Implementation Workflow
When implementing a task, follow this complete workflow:

### 1. Analysis Phase
- Read and understand the JIRA issue (if provided)
- Analyze existing codebase structure and patterns
- Identify files that need modification
- Plan the implementation approach

### 2. Implementation Phase
- Create/modify necessary files following established patterns
- Implement business logic in `internal/services/`
- Add/update handlers in `internal/api/handler/`
- Update models in `models/` if needed
- Follow dependency injection patterns using `samber/do`

### 3. Testing Phase
- Create comprehensive unit tests
- Mock external dependencies
- Ensure > 80% test coverage
- Run all tests to verify functionality

### 4. Documentation Phase
- Update Swagger annotations for API changes
- Update relevant documentation
- Ensure code comments are clear and helpful

### 5. PR Creation Phase
- Create branch with JIRA ID as name
- Commit changes with descriptive messages
- Push changes and create PR with formatted description
- Add required labels

## PR Labeling
- Always add the label `runtest:all:stable` to PRs created using GitHub CLI
- Add additional relevant labels based on the changes made

## MCP Server Configuration
- Use the atlassian-mcp-server for JIRA integration and issue management
- Leverage the server to fetch issue details, update status, and manage workflow
- Ensure all JIRA interactions follow proper authentication and authorization

## Code Quality Standards
- Follow Go best practices and conventions
- Use structured logging with Logrus
- Implement proper error handling with fiber.NewError
- Follow existing patterns for dependency injection
- Maintain backward compatibility with existing API clients
- Use interfaces for all external dependencies to enable mocking

## Prohibited Modifications
- Do not modify any `*.codegen.go` files
- Do not modify core configuration files without explicit requirements
- Do not change existing API contracts without proper versioning
- Do not modify database schema without migration strategies

## Environment Setup
- Go 1.23.6 is required
- Uses Fiber v2 web framework
- DynamoDB for data storage
- AWS services integration (S3, Secrets Manager)
- Docker containerization support

## Key Technologies and Patterns
- **Web Framework**: Fiber v2
- **Database**: DynamoDB with custom operations layer
- **Dependency Injection**: samber/do framework
- **Testing**: testify for assertions and mocks
- **Logging**: Logrus structured logging
- **Documentation**: Swagger/OpenAPI 3.0
- **Strategy Pattern**: Used for product-specific logic
- **License Validation**: Required for all protected endpoints

## Build and Development Commands
- `make all` - Build service with Swagger documentation
- `make build` - Build service only
- `make test` - Run all tests
- `make swagger` - Generate Swagger documentation
- `go test -race -vet=off ./...` - Run tests with race detection

## Service Architecture
This is the **Licensed Omnitruck API** service that provides license validation and entitlement checking for Chef product downloads. The service acts as a proxy to the Chef Omnitruck API with additional licensing controls.

### Core Components
- **API Handlers**: HTTP request processing and response formatting
- **Services Layer**: Business logic and orchestration
- **Clients Layer**: External API communication (Omnitruck, License, Replicated)
- **Strategy Pattern**: Product-specific implementation logic
- **Middleware**: License validation and database operations
- **Validators**: Input validation and business rule enforcement

### Database Tables
- `RelatedProductsTable` - Product relationships and dependencies
- `MetadataDetailsTable` - Package metadata and checksums
- `PackageManagersTable` - Available package manager configurations
- `PackageDetailsCurrentTable` - Current release package information
- `PackageDetailsStableTable` - Stable release package information

Remember to always follow the established patterns, maintain high test coverage, and ensure all changes go through the proper PR workflow with appropriate labeling and documentation.