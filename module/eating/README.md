# Eating module

This module is used to fully restore health for **only once**.

The purpose of this module is to let player fully restore health before using `kovojimas` module. `kovojimas` module only takes health into consideration *after* hitting the enemy, so if you start `kovojimas` without (some of) health and enemy hits you back - you die and lose 50% money. By using `eating` module prior `kovojimas`, you ensure that the first hit is always with full health.

After the first hit, `kovojimas` module will takes care of maintaining your health level. :)

## Example

```yaml
  - _module: eating    # This module ensures that you have 100% health before first hit
    food: UO10
  - _module: kovojimas # After each hit it maintains specified level of health
    vs: 1
    eating: UO10
    eating_threshold: 50
```
