# Vartai wait module

This module allows to wait (sleep) until vartai is open to fight.

By design, when vartai is not ready to be fighted, module always quits and routine goes to next module (which might mean that routine is starting over again). This `vartai_wait` module is intended to explicitly wait until vartai is open.

## Example

It's intended to be used before `vartai` module:

```yaml
  - _module: vartai_wait
  - _module: vartai
```
