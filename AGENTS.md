# Repository instructions

## Agent interoperability

For every change to agent instructions, tools, or protocol adapters, check the [AAIF project directory](https://aaif.io/projects) and follow the current upstream specification for each interface this repository implements. Use the AGENTS.md convention for repository guidance. Record the protocol revision and compatibility checks when an interface changes. Keep the relay collection contract compatible with the Foundry module and its MCP consumers. Test changes to collection routing in go-relay and build each affected frontend or documentation package.
