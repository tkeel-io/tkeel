# Developing TKeel

## Setup TKeel development environment

There are several options for getting an environment up and running for TKeel development:

- Use a [GitHub Codespace](https://docs.tkeel.io/contributing/codespaces/) configured for TKeel development. \[Requires [Beta sign-up](https://github.com/features/codespaces/signup)\]
- [Manually install](./setup-tkeel-development-env.md) the necessary tools and frameworks for developing TKeel on your device.

## Cloning the repo

```bash
cd $GOPATH/src
mkdir -p github.com/tkeel-io/tkeel
git clone https://github.com/tkeel-io/tkeel.git github.com/tkeel-io/tkeel
```

## Build the TKeel binaries

You can build TKeel binaries with the `make` tool.

> On Windows, the `make` commands must be run under [git-bash](https://www.atlassian.com/git/tutorials/git-bash).
>
> These instructions also require that a `make` alias has been created for `mingw32-make.exe` according to the [setup instructions](./setup-tkeel-development-env.md#installing-make).

- When running `make`, you need to be at the root of the `tkeel/tkeel` repo directory, for example: `$GOPATH/src/github.com/tkeel-io/tkeel`.

- Once built, the release binaries will be found in `./dist/{os}_{arch}/release/`, where `{os}_{arch}` is your current OS and architecture.

  For example, running `make build` on an Intel-based MacOS will generate the directory `./dist/darwin_amd64/release`.

- To build for your current local environment:

   ```bash
   cd $GOPATH/src/github.com/tkeel-io/tkeel/
   make build
   ```

- To cross-compile for a different platform:

   ```bash
   make build GOOS=windows GOARCH=amd64
   ```

  For example, developers on Windows who prefer to develop in [WSL2](https://docs.microsoft.com/en-us/windows/wsl/install-win10) can use the Linux development environment to cross-compile binaries like `tkeeld.exe` that run on Windows natively.
