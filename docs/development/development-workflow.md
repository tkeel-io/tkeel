# Development Workflow

The following graph shows the workflow how to develop TKeel backend.

## Step 1. Fork

1. Visit https://github.com/tkeel-io/keel
2. Click `Fork` button to create a fork of the project to your GitHub account.

## Step 2. Clone fork to local storage

Per Go's [workspace instructions](https://golang.org/doc/code.html#Workspaces), place TKeel code on your `GOPATH` using the following cloning procedure.

Define a local working directory:

```bash
export working_dir=$GOPATH/src/tkeel.io
export user={your github profile name}
```

Create your clone locally:

```bash
mkdir -p $working_dir
cd $working_dir
git clone https://github.com/$user/keel.git
cd $working_dir/keel
git remote add upstream https://github.com/tkeel-io/keel.git

# Never push to upstream master
git remote set-url --push upstream no_push

# Confirm your remotes make sense:
git remote -v
```

## Step 3. Keep your branch in sync

```bash
git fetch upstream
git checkout master
git rebase upstream/master
```

## Step 4. Add new features or fix issues

Create a branch from master:

```bash
git checkout -b myfeature
```

Then edit code on the `myfeature` branch. You can refer to [effective_go](https://golang.org/doc/effective_go.html) while writing code.

### Test and build

Currently, the make rules only contain simple checks such as vet, unit test, will add e2e tests soon.

### Run and test

```bash
make all
# Run every unit test
make test
```

Run `make help` for additional information on these make targets.

## Step 5. Development in new branch

### Sync with upstream

After the test is completed, it is a good practice to keep your local in sync with upstream to avoid conflicts.

```bash
# Rebase your master branch of your local repo.
git checkout master
git rebase upstream/master

# Then make your development branch in sync with master branch
git checkout new_feature
git rebase -i master
```

### Commit local changes

```bash
git add <file>
git commit -s -m "add your description"
```

## Step 6. Push to your fork

When ready to review (or just to establish an offsite backup of your work), push your branch to your fork on GitHub:

```bash
git push -f ${your_remote_name} myfeature
```

## Step 7. Create a PR

- Visit your fork at https://github.com/$user/keel
- Click the` Compare & Pull Request` button next to your myfeature branch.
- Check out the [pull request process](pull-request.md) for more details and advice.
