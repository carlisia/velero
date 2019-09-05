# Plan to extract the provider plugins out of (the Velero) tree

Currently, the Velero project contains primitive plugins for three cloud providers: AWS, Azure, and GCP. The Velero team has decided to extract each of those plugins into their own separate repository.  This document details the steps necessary to create the new repositories, as well as a general design for what each plugin project will look like.

## Goals

- Have 3 new repositories for each cloud provider plugin currently supported by the Velero team: AWS, Azure, and GCP
- Have the currently primitive cloud provider plugins behave like any other plugin external to Velero

## Non Goals

- Extend the Velero plugin framework capability in any way
- Create GH repositories for any plugin other then the currently primitive plugins
- Extract out any plugin that is not a cloud provider plugin (ex: item action related plugins)

## Background

With more and more providers wanting to support Velero, it gets more difficult to justify excluding those from being in-tree just as with the three original ones. At the same time, if we were to include any more plugins in-tree, it would ultimately become the responsibility of the Velero team to maintain all plugins. This move aims to equalize the field so all plugins are treated equally. We also hope that, with time, developers interested in getting involved in the upkeep of those plugins will become active enough to be promoted to maintainers. Lastly, having the plugins live in their own individual repositories allows for iteration on them separately from the core codebase.

## Action items

### Todo list

- [ ] Use GH UI to create each repository in the new VMW org. Who: new org owner; TBD
- [ ] Make owners of the Velero repo owners of each repo in the new org. Who: new org owner; TBD
- [ ] Create deployment and grpc-push scripts with the new location path. Who: @carlisia
- [ ] Add Travis CI. Who: Any of the new repo owners; TBD
- [ ] Add webhook: travis CI. Who: Any of the new repo owners; TBD

### Notes/How-Tos

#### Creating the GH repository

[Pending] We will find out this week who will be the organization owner(s) who will accept this transfer in the new GH org. This organization owner will make all current owners in the Velero repo owners in each of the new org plugin repos.

#### Setting up Travis CI

Someone with owner permission on the new repository needs to go to their Travis CI account and authorize Travis CI on the repo. Here are instructions: https://docs.travis-ci.com/user/tutorial/.

After this, any webhook notifications can be added following these instructions: https://docs.travis-ci.com/user/notifications/#configuring-webhook-notifications.

## High-Level Design

Each provider plugin will be an independent project, using the Velero library to implement their specific functionalities.

The way Velero is installed will be changed to accomodate installing these plugins at deploy time, namely the Velero `install` command, as well as the Helm chart.

Each plugin repository will need to have their respective images built and pushed to the same registry as the Velero images.

One to two paragraphs that describe the high level changes that will be made to implement this proposal.

## Detailed Design

=======

Enhance velero install do plugin installation at deploy time, and make sure it supports all plugins.
Enhance the Helm chart to support plugin installation at deploy time.
have a separate repo per provider -- that way each binary/image can be as small as possible and limited to a single provider's SDK.

what's the repo structure? (repo per provider, etc)
build/release process changes?
where do docs live?
are there any places outside the plugins where we depend on the cloud-provider SDKs? can we eliminate those dependencies too?
what does the upgrade path look like for users currently using the in-tree plugins?
each separate repo for plugins will need to push to the same image repo

----
De we want to keep the logic that adds velero.io to any plugin name that isn't namespaced with this change?
Example: https://github.com/heptio/velero/blob/master/pkg/plugin/clientmgmt/manager.go#L129-L130

----
Another question: are we moving just the cloud provider plugins out? I think some plugins, such as the pod -> pvc -> pv backupitemaction ones, make sense to stay in the core repo as they provide some important logic that just happens to be implemented in a plugin.

I personally don't really see any reason to move the item action plugins out, interested to hear others' thoughts.


=======

A detailed design describing how the changes to the product should be made.

The names of types, fields, interfaces, and methods should be agreed on here, not debated in code review.
The same applies to changes in CRDs, YAML examples, and so on.

Ideally the changes should be made in sequence so that the work required to implement this design can be done incrementally, possibly in parallel.

## Alternatives Considered

If there are alternative high level or detailed designs that were not pursued they should be called out here with a brief explanation of why they were not pursued.

## Security Considerations

If this proposal has an impact to the security of the product, its users, or data stored or transmitted via the product, they must be addressed here.
