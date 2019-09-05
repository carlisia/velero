# Design proposal template (replace with your proposal's title)

One to two sentences that describes the goal of this proposal.
The reader should be able to tell by the title, and the opening paragraph, if this document is relevant to them.

_Note_: The preferred style for design documents is one sentence per line.
*Do not wrap lines*.
This aids in review of the document as changes to a line are not obscured by the reflowing those changes caused and has a side effect of avoiding debate about one or two space after a period.

## Goals

- A short list of things which will be accomplished by implementing this proposal.
- Two things is ok.
- Three is pushing it.
- More than three goals suggests that the proposal's scope is too large.

## Non Goals

- A short list of items which are:
- a. out of scope
- b. follow on items which are deliberately excluded from this proposal.

## Background

One to two paragraphs of exposition to set the context for this proposal.

## High-Level Design

One to two paragraphs that describe the high level changes that will be made to implement this proposal.

## Action items

### Todo list

- [ ] Use GH UI to create each repository in the new VMW org. Who: new org owner; TBD
- [ ] Make owners of the Velero repo owners of each repo in the new org. Who: new org owner; TBD
- [ ] Create deployment and grpc-push scripts with the new location path. Who: @carlisia
- [ ] Add Travis CI. Who: Any of the new repo owners; TBD
- [ ] Add webhook: signoff checker. Who: Any of the new repo owners; TBD
- [ ] Add webhook: travis CI. Who: Any of the new repo owners; TBD

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
