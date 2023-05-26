# OpenAPI

We're using `swaggo` in SPC to generate the OpenAPI 2.0 specification. There's work in progress to update swaggo to OpenAPI 3.x but it's very slow going.

swaggo is code -> spec. It's the only viable option, there's no good alternative for code -> spec, at least in Go land. I investigate the other, seems more popular route, spec -> code.

https://github.com/deepmap/oapi-codegen is the de facto recommendation for Go. There is `Gin` support, but this is not their default or preferred router. The team prefer to use `echo`.

Findings as of 2023-05-06:

* oapi-codegen is a mature and solid project
* Gin support is adequate, works and has various customisation
* However, it's missing a crucial feature that we use: the ability to have **multiple** handlers per route. We use this so we can add custom rate limiter per path and calling other functions for that path.
* `swaggo` has the full flexibility but it's code -> spec so our spec is often incorrect.
* `swaggo` also only suppport OpenAPI 2.0


For now, we're sticking with `swaggo`. Despite only up to OpenAPI 2.0 this is our best option for now.
