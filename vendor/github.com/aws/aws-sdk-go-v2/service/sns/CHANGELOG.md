# v1.29.1 (2024-02-23)

* **Bug Fix**: Move all common, SDK-side middleware stack ops into the service client module to prevent cross-module compatibility issues in the future.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.29.0 (2024-02-22)

* **Feature**: Add middleware stack snapshot tests.

# v1.28.2 (2024-02-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.28.1 (2024-02-20)

* **Bug Fix**: When sourcing values for a service's `EndpointParameters`, the lack of a configured region (i.e. `options.Region == ""`) will now translate to a `nil` value for `EndpointParameters.Region` instead of a pointer to the empty string `""`. This will result in a much more explicit error when calling an operation instead of an obscure hostname lookup failure.

# v1.28.0 (2024-02-16)

* **Feature**: This release marks phone numbers as sensitive inputs.

# v1.27.0 (2024-02-13)

* **Feature**: Bump minimum Go version to 1.20 per our language support policy.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.7 (2024-01-04)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.6 (2023-12-20)

* No change notes available for this release.

# v1.26.5 (2023-12-08)

* **Bug Fix**: Reinstate presence of default Retryer in functional options, but still respect max attempts set therein.

# v1.26.4 (2023-12-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.3 (2023-12-06)

* **Bug Fix**: Restore pre-refactor auth behavior where all operations could technically be performed anonymously.

# v1.26.2 (2023-12-01)

* **Bug Fix**: Correct wrapping of errors in authentication workflow.
* **Bug Fix**: Correctly recognize cache-wrapped instances of AnonymousCredentials at client construction.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.1 (2023-11-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.26.0 (2023-11-29)

* **Feature**: Expose Options() accessor on service clients.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.5 (2023-11-28.2)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.4 (2023-11-28)

* **Bug Fix**: Respect setting RetryMaxAttempts in functional options at client construction.

# v1.25.3 (2023-11-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.2 (2023-11-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.1 (2023-11-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.25.0 (2023-11-01)

* **Feature**: Adds support for configured endpoints via environment variables and the AWS shared configuration file.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.24.0 (2023-10-31)

* **Feature**: **BREAKING CHANGE**: Bump minimum go version to 1.19 per the revised [go version support policy](https://aws.amazon.com/blogs/developer/aws-sdk-for-go-aligns-with-go-release-policy-on-supported-runtimes/).
* **Dependency Update**: Updated to the latest SDK module versions

# v1.23.0 (2023-10-26)

* **Feature**: Message Archiving and Replay is now supported in Amazon SNS for FIFO topics.

# v1.22.2 (2023-10-12)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.1 (2023-10-06)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.22.0 (2023-09-18)

* **Announcement**: [BREAKFIX] Change in MaxResults datatype from value to pointer type in cognito-sync service.
* **Feature**: Adds several endpoint ruleset changes across all models: smaller rulesets, removed non-unique regional endpoints, fixes FIPS and DualStack endpoints, and make region not required in SDK::Endpoint. Additional breakfix to cognito-sync field.

# v1.21.5 (2023-08-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.4 (2023-08-18)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.3 (2023-08-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.2 (2023-08-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.21.1 (2023-08-01)

* No change notes available for this release.

# v1.21.0 (2023-07-31)

* **Feature**: Adds support for smithy-modeled endpoint resolution. A new rules-based endpoint resolution will be added to the SDK which will supercede and deprecate existing endpoint resolution. Specifically, EndpointResolver will be deprecated while BaseEndpoint and EndpointResolverV2 will take its place. For more information, please see the Endpoints section in our Developer Guide.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.15 (2023-07-28)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.14 (2023-07-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.13 (2023-06-15)

* No change notes available for this release.

# v1.20.12 (2023-06-13)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.11 (2023-05-08)

* No change notes available for this release.

# v1.20.10 (2023-05-04)

* No change notes available for this release.

# v1.20.9 (2023-04-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.8 (2023-04-10)

* No change notes available for this release.

# v1.20.7 (2023-04-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.6 (2023-03-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.5 (2023-03-10)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.4 (2023-02-22)

* **Bug Fix**: Prevent nil pointer dereference when retrieving error codes.

# v1.20.3 (2023-02-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.20.2 (2023-02-10)

* **Documentation**: This release adds support for SNS X-Ray active tracing as well as other updates.

# v1.20.1 (2023-02-03)

* **Dependency Update**: Updated to the latest SDK module versions
* **Dependency Update**: Upgrade smithy to 1.27.2 and correct empty query list serialization.

# v1.20.0 (2023-02-01)

* **Feature**: Additional attributes added for set-topic-attributes.

# v1.19.1 (2023-01-23)

* No change notes available for this release.

# v1.19.0 (2023-01-05)

* **Feature**: Add `ErrorCodeOverride` field to all error structs (aws/smithy-go#401).

# v1.18.8 (2022-12-15)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.7 (2022-12-02)

* **Documentation**: This release adds the message payload-filtering feature to the SNS Subscribe, SetSubscriptionAttributes, and GetSubscriptionAttributes API actions
* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.6 (2022-11-22)

* No change notes available for this release.

# v1.18.5 (2022-11-16)

* No change notes available for this release.

# v1.18.4 (2022-11-10)

* No change notes available for this release.

# v1.18.3 (2022-10-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.2 (2022-10-21)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.1 (2022-09-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.18.0 (2022-09-14)

* **Feature**: Amazon SNS introduces the Data Protection Policy APIs, which enable customers to attach a data protection policy to an SNS topic. This allows topic owners to enable the new message data protection feature to audit and block sensitive data that is exchanged through their topics.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.17 (2022-09-02)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.16 (2022-08-31)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.15 (2022-08-30)

* No change notes available for this release.

# v1.17.14 (2022-08-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.13 (2022-08-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.12 (2022-08-09)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.11 (2022-08-08)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.10 (2022-08-01)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.9 (2022-07-05)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.8 (2022-06-29)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.7 (2022-06-07)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.6 (2022-05-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.5 (2022-04-25)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.4 (2022-03-30)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.3 (2022-03-24)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.2 (2022-03-23)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.17.1 (2022-03-08.2)

* No change notes available for this release.

# v1.17.0 (2022-03-08)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.16.0 (2022-02-24)

* **Feature**: API client updated
* **Feature**: Adds RetryMaxAttempts and RetryMod to API client Options. This allows the API clients' default Retryer to be configured from the shared configuration files or environment variables. Adding a new Retry mode of `Adaptive`. `Adaptive` retry mode is an experimental mode, adding client rate limiting when throttles reponses are received from an API. See [retry.AdaptiveMode](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#AdaptiveMode) for more details, and configuration options.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.15.0 (2022-01-14)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.14.0 (2022-01-07)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.13.0 (2021-12-21)

* **Feature**: API Paginators now support specifying the initial starting token, and support stopping on empty string tokens.
* **Feature**: Updated to latest service endpoints

# v1.12.1 (2021-12-02)

* **Bug Fix**: Fixes a bug that prevented aws.EndpointResolverWithOptions from being used by the service client. ([#1514](https://github.com/aws/aws-sdk-go-v2/pull/1514))
* **Dependency Update**: Updated to the latest SDK module versions

# v1.12.0 (2021-11-19)

* **Feature**: API client updated
* **Dependency Update**: Updated to the latest SDK module versions

# v1.11.0 (2021-11-12)

* **Feature**: Service clients now support custom endpoints that have an initial URI path defined.

# v1.10.0 (2021-11-06)

* **Feature**: The SDK now supports configuration of FIPS and DualStack endpoints using environment variables, shared configuration, or programmatically.
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.9.0 (2021-10-21)

* **Feature**: API client updated
* **Feature**: Updated  to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.2 (2021-10-11)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.1 (2021-09-17)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.8.0 (2021-08-27)

* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.2 (2021-08-19)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.1 (2021-08-04)

* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version.
* **Dependency Update**: Updated to the latest SDK module versions

# v1.7.0 (2021-07-15)

* **Feature**: The ErrorCode method on generated service error types has been corrected to match the API model.
* **Documentation**: Updated service model to latest revision.
* **Dependency Update**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.6.0 (2021-06-25)

* **Feature**: API client updated
* **Feature**: Updated `github.com/aws/smithy-go` to latest version
* **Dependency Update**: Updated to the latest SDK module versions

# v1.5.0 (2021-06-04)

* **Feature**: Updated service client to latest API model.

# v1.4.1 (2021-05-20)

* **Dependency Update**: Updated to the latest SDK module versions

# v1.4.0 (2021-05-14)

* **Feature**: Constant has been added to modules to enable runtime version inspection for reporting.
* **Dependency Update**: Updated to the latest SDK module versions

