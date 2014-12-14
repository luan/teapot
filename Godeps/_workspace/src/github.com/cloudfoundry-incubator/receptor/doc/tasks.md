# Tasks

Diego can run one-off work in the form of Tasks.  When a Task is submitted Diego allocates resources on a Cell, runs the Task, and then reports on the Task's results.  Tasks are guaranteed to run at most once.

## Describing Tasks

When submitting a Task you `POST` a valid `TaskCreateRequest`.  The [API reference](api_tasks.md) includes the details of the request.  Here we simply describe what goes into a `TaskCreateRequest`:

```
{
    "task_guid": "some-guid",
    "domain": "some-domain",

    "stack": "lucid64",

    "root_fs": "docker:///docker-org/docker-image",
    "env": [
        {"name": "ENV_NAME_A", "value": "ENV_VALUE_A"},
        {"name": "ENV_NAME_B", "value": "ENV_VALUE_B"}
    ],

    "cpu_weight": 57,
    "disk_mb": 1024,
    "memory_mb": 128,

    "action":  ACTION (see below),

    "result_file": "/path/to/return",
    "completion_callback_url": "http://optional/callback/url",

    "log_guid": "some-log-guid",
    "log_source": "some-log-source",

    "annotation": "arbitrary metadata"
}
```

Let's describe each of these fields in turn.

#### Task Identifiers

#### `task_guid` [required]

It is up to the consumer of Diego to provide a *globally unique* `task_guid`.  To subsequently fetch the Task you refer to it by its `task_guid`.

- It is an error to attempt to create a Task whose `task_guid` matches that of an existing Task.
- The `task_guid` must only include the characters `a-z`, `A-Z`, `0-9`, `_` and `-`.
- The `task_guid` must not be empty

#### `domain` [required]

The consumer of Diego may organize their Tasks into groupings called Domains.  These are purely organizational (e.g. for enabling multiple consumers to use Diego without colliding) and have no implications on the Task's placement or lifecycle.  It is possible to fetch all Tasks in a given Domain.

- It is an error to provide an empty `domain`.

#### Task Placement

In the future Diego will support the notion of Placement Pools via arbitrary tags associated with Cells.  For now, this functionality is limited to the notion of `stack`.

#### `stack` [required]

Diego can support different target platforms (linux, windows, etc.). `stack` allows you to select which target platform the Task must run against.  For a typical Diego deployment you should set `stack` to `lucid64`

- It is an error to provide an empty `stack`.

#### Container Contents and Environment

#### `root_fs` [optional]

By default, when provisioning a container, Diego will mount a pre-configured root filesystem.  Currently, the default filesystem provided by [diego-release](https://github.com/cloudfoundry-incubator/diego-release) is based on lucid64 and is geared towards supporting the Cloud Foundry buildpacks.

It is possible, however, to provide a custom root filesystem by specifying a Dockerimage for `root_fs`:

```
"root_fs": "docker:///docker-org/docker-image#docker-tag"
```

Currently, only the public docker hub is supported.

> You *must* specify the dockerimage `root_fs` uri as specified, including the leading `docker:///`!

> [Diego-Edge](http://github.com/cloudfoundry-incubator/diego-lite) does not ship with a default rootfs.  You must specify a docker-image when using Diego-Edge.  You can mount the filesystem provided by diego-release by specifying `"root_fs": "docker:///cloudfoundry/lucid64"` or `"root_fs": "docker:///cloudfoundry/trusty64"`.

#### `env` [optional]

Diego supports the notion of container-level environment variables.  All processes that run in the container will inherit these environment variables.

For more details on the environment variables provided to processes in the container, read [Container Runtime Environment](environment.md)

#### Container Limits

#### `cpu_weight` [optional]

To control the CPU shares provided to a container, set `cpu_weight`.  This must be a positive number in the range `1-100`.  The `cpu_weight` enforces a relative fair share of the CPU among containers.  It's best explained with examples.  Consider the following scenarios (we shall assume that each container is running a busy process that is attempting to consumer as many CPU resources as possible):

- Two containers, with equal values of `cpu_weight`: both containers will receive equal shares of CPU time.
- Two containers, one with `cpu_weight=50` the other with `cpu_weight=100`: the latter will get (roughly) 2/3 of the CPU time, the former 1/3.

#### `disk_mb` [optional]

A disk quota applied to the entire container.  Any data written on top of the RootFS counts against the Disk Quota.  Processes that attempt to exceed this limit will not be allowed to write to disk.

- `disk_mb` must be an integer > 0
- The units are megabytes

#### `memory_mb` [optional]

A memory limit applied to the entire container.  If the aggregate memory consumption by all processs running in the container exceeds this value, the container will be destroyed.

- `memory_mb` must be an integer > 0
- The units are megabytes

#### Actions

#### `action` [required]

Encodes the action to run when running the Task.  For more details see [actions](actions.md)

#### Task Completion and Output

When the `action` on a Task terminates the Task is marked as `COMPLETED`.

#### `result_file` [optional]

When a Task completes succesfully Diego can fetch and return the contents of a file in the container.  This is made available in the `result` field of the `TaskResponse` (see [below](#retreiving-tasks)).

To do this, set `result_file` to a valid path in the container.

- Diego only returns the first 10KB of the `result_file`.  If you need to communicate back larger datasets, consider using an `UploadAction` to upload the result file to a blob store.

#### `completion_callback_url` [optional]

Consumers of Diego have two options to learn that a Task has `COMPLETED`: they can either poll the action or register a callback.

If a `completion_callback_url` is provided Diego will `POST` to the provided URL as soon as the Task completes.  The body of the `POST` will include the `TaskResponse` (see [below](#retreiving-tasks)).

- Any response from the callback (be it success or failure) will resolve the Task (removing it from Diego).  
- However, if the callback responds with `503` or `504` Diego will immediately retry the callback up to 3 times.  If the `503/504` status persists Diego will try again after a period of time (typically within ~30 seconds).
- If the callback times out or a connection cannot be established, Diego will try again after a period of time (typically within ~30 seconds).
- Diego will eventually (after ~2 minutes) give up on the Task if the callback does not respond succesfully.

#### Logging

Diego uses [doppler](https://github.com/cloudfoundry-incubator/doppler) to emit logs generated by container processes to the user.

#### `log_guid` [optional]

`log_guid` controls the doppler guid associated with logs coming from Task processes.  One typically sets the `log_guid` to the `task_guid` though this is not strictly necessary.

#### `log_source` [optional]

`log_source` is an identifier emitted with each log line.  Individual `RunAction`s can override the `log_source`.  This allows a consumer of the log stream to distinguish between the logs of different processes.

#### Attaching Arbitrary Metadata

#### `annotation` [optional]

Diego allows arbitrary annotations to be attached to a Task.  The annotation must not exceed 10 kilobytes in size.

## Retreiving Tasks

To learn that a Task is completed you must either register a `completion_callback_url` or periodically poll the API to fetch the Task in question.  In both cases, you will receive an object that includes **all the fields on the `TaskCreateRequest`** and the following additional fields:

```
{
    ... all TaskCreateRequest fields...

    "state": "RUNNING",
    
    "cell_id": "cell-identifier",

    "failed": true/false,
    "failure_reason": "why it failed",
    "result": "the contents of result_file",
}
```

Let's describe each of these fields in turn.

#### `state`

Tasks travel through a series of state transitions throughout their lifecycle.  These are described in [The Task Lifecycle](#the-task-lifecycle) below.

`state` will be a string and one of `INVALID`, `PENDING`, `CLAIMED`, `RUNNING`, `COMPLETED`, `RESOLVING`.

#### `cell_id`

Once claimed, a Task will include the ID of the Diego cell it is running on.

#### `failed`

Once a Task enters the `COMPLETED` state, `failed` will be a boolean indicating whether the Task completed succesfully or unsuccesfully.

#### `failure_reason`

If `failed` is `true`, `failure_reason` will be a short string indicating why the Task failed.  Sometimes, in the case of a `RunAction` that has failed this will simply read (e.g.) `exit status 1`.  To debug the Task you will need to fetch the logs from doppler.

#### `result`

If `result_file` was specified and the Task has completed succesfully, `result` will include the first 10KB of the `result_file`.

## The Task lifecycle

Tasks in Diego undergo a simple lifecycle encoded in the Tasks's state:

- When first created a Task enters the `PENDING` state.
- When succesfully allocated to a Diego Cell the Task will enter the `CLAIMED` state.  At this point the Task's `cell_id` will be populated.
- When the Cell begins to create the container and run the Task action, the Task enters the `RUNNING` state.
- When the Task completes, the Cell annotates the `TaskResponse` with `failed`, `failure_reason`, and `result`, and puts the Task in the `COMPLETED` state.

At this point it is up to the consumer of Diego to acknowledge and resolve the completed Task.  This can either be done via a completion callback (described [above](#completion_callback_url)) or by [deleting](delete_tasks.md) the Task.  When the Task is being resolved it first enters the `RESOLVING` state and is ultimately removed from Diego.

Diego will automatically reap Tasks that remain unresolved after 2 minutes.

> The `RESOLVING` state exists to ensure that the `completion_callback_url` is initially called at most once per Task.

> There are a variety of timeouts associated with the `PENDING` and `CLAIMED` states.  It is possible for a Task to jump directly from `PENDING` or `CLAIMED` to `COMPLETED` (and `failed`) if any of these timeouts expire.  If you would like to impose a time limit on how long the Task is allowed to run you can use a `TimeoutAction`.

## Cancelling Tasks

Diego supports cancelling inflight tasks.  More documentation on this is available [here](cancel_tasks.md).

[back](README.md)