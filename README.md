# tsc-resolution-tracker

Explains what modules TypeScript is including and why.

This is a thin wrapper over:

```
tsc --traceResolution
```

Let us create a TypeScript project with `pulumi new aws-typescript`.


## What modules are imported

```
GOSUMDB=off go run github.com/t0yv0/tsc-resolution-tracker@v0.2.0 | head
index.ts -> @pulumi/pulumi/index.d.ts
index.ts -> @pulumi/aws/index.d.ts
index.ts -> @pulumi/awsx/index.d.ts
@pulumi/pulumi/index.d.ts -> source-map-support/register.js
@pulumi/pulumi/index.d.ts -> @pulumi/pulumi/config.d.ts
@pulumi/pulumi/index.d.ts -> @pulumi/pulumi/errors.d.ts
@pulumi/pulumi/index.d.ts -> @pulumi/pulumi/invoke.d.ts
@pulumi/pulumi/index.d.ts -> @pulumi/pulumi/metadata.d.ts
@pulumi/pulumi/index.d.ts -> @pulumi/pulumi/output.d.ts
@pulumi/pulumi/index.d.ts -> @pulumi/pulumi/resource.d.ts
```


## Why is this module imported

```
GOSUMDB=off go run github.com/t0yv0/tsc-resolution-tracker@v0.2.0 -why '@pulumi/aws/lambda/permission.d.ts'
index.ts ->
@pulumi/aws/index.d.ts ->
@pulumi/aws/arn.d.ts ->
@pulumi/aws/iam/index.d.ts ->
@pulumi/aws/iam/getPolicyDocument.d.ts ->
@pulumi/aws/types/input.d.ts ->
@pulumi/aws/s3/index.d.ts ->
@pulumi/aws/s3/s3Mixins.d.ts ->
@pulumi/aws/lambda/index.d.ts ->
@pulumi/aws/lambda/permission.d.ts

total modules matching "@pulumi/aws/lambda/permission.d.ts": 1
```
