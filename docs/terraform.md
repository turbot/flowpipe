

Terraform moved all Terraform's non-library package to internal so we can't reuse them anymore.

https://github.com/hashicorp/terraform/commit/dc0ccb9c3aeaa243d498da2d87bc6b7e90d1260e

```
Move addrs/ to internal/addrs/
This is part of a general effort to move all of Terraform's non-library
package surface under internal in order to reinforce that these are for
internal use within Terraform only.

If you were previously importing packages under this prefix into an
external codebase, you could pin to an earlier release tag as an interim
solution until you've make a plan to achieve the same functionality some
other way.
```
