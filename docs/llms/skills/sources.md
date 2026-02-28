# Sources

Research citations for GO_DEVELOPMENT.md and GO_TESTING.md skills.

## Authoritative Sources

### Official Go Documentation
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) - Official wiki on common code review feedback
- [Google Go Style Guide](https://google.github.io/styleguide/go/) - Google's internal Go style guide (public version)

### Industry Style Guides
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) - Comprehensive style guide from Uber's Go team

### Production Experience
- [Peter Bourgon: Go Best Practices for Production](https://peter.bourgon.org/go-in-production/) - Battle-tested patterns from SoundCloud's Go infrastructure
- [Google SRE: How Google Uses Go](https://opensource.googleblog.com/2021/04/actuating-google-production-how-googles-sre-team-uses-go.html) - Google's production Go usage

### Testing
- [Three Dots Labs: Database Integration Testing](https://threedots.tech/post/database-integration-testing/) - Practical principles for high-quality integration tests
- [Speedscale: Golang Testing Frameworks](https://speedscale.com/blog/golang-testing-frameworks-for-every-type-of-test/) - Framework comparison and best practices

## Key Principles Extracted

### From Go Code Review Comments
- Error handling: check errors, indent error flow, lowercase error strings
- Naming: short names for short scopes, avoid package name in identifiers
- Interfaces: define at consumer, not producer; don't define before use
- Receivers: consistent pointer vs value, don't mix

### From Uber Style Guide
- Handle errors once (don't log and return)
- Verify interface compliance at compile time
- Prefer strconv over fmt for performance
- Pre-allocate slices and maps when size is known

### From Peter Bourgon
- Define flags in main() to enforce dependency injection
- Log only actionable information
- Use build tags for integration tests
- "Don't create structure until you demonstrably need it"

### From Three Dots Labs
- Tests must be fast (<10s locally)
- Avoid time.Sleep - use synchronization
- Sabotage test: break implementation, verify test catches it
- 70-80% coverage distributed across unit/integration/component

## Date Reviewed
2024-12-23
