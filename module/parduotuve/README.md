# Parduotuve module

This module allows to buy/sell items in main shop.

## Examples

Buy max possible amount of items

```yaml
  - _module: parduotuve
    action: pirkti
    item: MA1
```

Same as above

```yaml
  - _module: parduotuve
    action: pirkti
    item: MA1
    amount: 0
```

Buy exactly 5 items

```yaml
  - _module: parduotuve
    action: pirkti
    item: MA1
    amount: 5
```

Buy max-1 amount of items (so you will have 1 free space in inventory afterwards)

```yaml
  - _module: parduotuve
    action: pirkti
    item: MA1
    amount: -1
```

Sell max amount of items

```yaml
  - _module: parduotuve
    action: parduoti
    item: MA1
```

Sell max-1 amount of items (1 item will remain in inventory, the rest of the items will be sold)

```yaml
  - _module: parduotuve
    action: parduoti
    item: MA1
    amount: 0
```
