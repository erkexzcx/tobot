# Slayer module

This module allows to perform actions in "Kovu laukas" with optional Slayer contracts.

## Example

```yaml
  - _module: slayer
    _count: 240
    vs: 1     # ID of enemy which one to fight. Key in URL is 'vs'.
    slayer: 1 # OPTIONAL - ID of Slayer contract which one to fight. Key in URL is 'nr'. If you don't pick it, then only simple fighting will be performed.
```
