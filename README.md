<h1 align="center">Welcome to Cron-GQL 👋</h1>
<p>
</p>

> Go-based cron scheduler wrapped in a GraphQL interface

## How to use

### 1. Download code

Clone the repository:

```
git clone git@github.com:henry74/cron-gql.git
```

### 2. Start the GraphQL server

```sh
go run server/server.go
```

OR

```sh
go build -o bin/cron-gql server/server.go
./bin/cron-gql
```

### 3. Using the GraphQL API

Navigate to [http://localhost:8080](http://localhost:8080) in your browser to explore the API of your GraphQL server in a [GraphQL Playground](https://github.com/prisma/graphql-playground).

The schema that specifies the API operations of your GraphQL server is defined in [`./schema.graphql`](./schema.graphql). Below are a number of operations that you can send to the API using the GraphQL Playground.

Feel free to adjust any operation by adding or removing fields. The GraphQL Playground helps you with its auto-completion and query validation features.

#### Add a new scheduled job

```graphql
mutation {
  addJob(
    input: {
      cronExp: "0 * * * *"
      rootDir: "/home/henry"
      cmd: "ls"
      args: "-alh"
      tags: ["directory", "hourly"]
    }
  ) {
    jobID
    cronExp
    cmd
    args
    tags
  }
}
```

<Details><Summary><strong>See more API operations</strong></Summary>

#### List all scheduled jobs

```graphql
query {
  jobs {
    jobID
    cronExp
    cmd
    args
    tags
    lastScheduledRun
    nextScheduledRun
  }
}
```

> **Note**: You can filter jobs by a specific `jobID` or list of `tags` e.g. using `jobs(input: { tags:["hourly"] }) {...}`.

#### Force a scheduled job to run immediately

```graphql
mutation {
  runJob(JobID: 1) {
    jobID
    cronExp
    cmd
    args
    tags
    lastScheduledRun
    nextScheduledRun
    lastForcedRun
  }
}
```

#### Remove a scheduled job

```graphql
mutation {
  removeJob(JobID: 1) {
    jobID
    cronExp
    cmd
    args
    tags
    lastScheduledRun
    nextScheduledRun
    lastForcedRun
  }
}
```

</Details>

### 4. Changing the GraphQL schema

After you made changes to `schema.graphql`, you need to update the generated types in `./generated.go` and potentially also adjust the resolver implementations in `./resolver.go`:

```sh
go generate ./... # from root
```

This updates `./generated.go` to incorporate the schema changes in your Go type definitions.

## Author

👤 **Henry H**

- Github: [@henry74](https://github.com/henry74)

## Show your support

Give a ⭐️ if this project helped you!

---

MIT License

Copyright (c) 2019 Henry Hwangbo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
