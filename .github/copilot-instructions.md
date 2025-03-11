### Update API spec

- Always update API spec under api/openapi.yaml if there are any Go code changes in the backend/ directory.

### General

- For any changes you make, summarize in the `.copilot-changelog.md` file.

### Run Tests

- Always try to run the unit tests using `make unit-test` after you are done with any code changes.

### Build binary

Always try check if binary creation works using `make build` after you are done with any code changes.
