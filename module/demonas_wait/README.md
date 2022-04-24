# Demonas wait module

This module allows to wait (sleep) until Demonas is open to fight.

By design, when demonas is not ready to be fighted, module always quits and routine goes to next module (which might mean that routine is starting over again). This `demonas_wait` module is intended to explicitly wait until demonas can be fought.

## Example

It's intended to be used before `demonas` module:

```yaml
  - _module: demonas_wait
  - _module: demonas
```
