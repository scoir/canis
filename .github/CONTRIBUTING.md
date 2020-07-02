# How to contribute

*We are still working on improving these guidelines.*

Thank you for taking the time to contribute to `scoir/canis`!

Contributing:
* Submit a [bug report](#bug-report)
* Submit a [pull request](#pull-request)

## Bug Report

Did you find a bug? Be sure to set a clear and concise title. Do your best to
include a code sample that illustrates the test case. Use the
[template](ISSUE_TEMPLATE.md).

## Pull Request

Use the [template](PULL_REQUEST_TEMPLATE.md) and make sure:

* **Required:** The build must pass. In particular, the following build targets
must pass successfully:
  * `make android-aar`
  * `make ios-framework`
* **Required:** Adherence to the  [Developer Certificate of Origin
(DCO)](https://developercertificate.org/) version 1.1 (`git --signoff`).
* **Required:** *squash your commits*. Yes, we know - it's nice to be able to
rollback every single change you make while troubleshooting or when requested to
exclude a subset or your change. The problem is the project's history tends to
become polluted with useless commit messages such as "removed trailing spaces".
It also makes it harder to revert specific changes when they are spread out like
this. We care about preserving our project's commit history in a usable state,
and as such, we politely request that you deliver your changes in a single
commit.
* Number of lines changed should not exceed 500. We're reasonable people - if
your PR is just a *little* over the top then we might still merge it. Such cases
are exceptional and are handled on a case-by-case basis.

The contents of this file are heavily "borrowed" from  aries-framework-go, mainly because it is are good.