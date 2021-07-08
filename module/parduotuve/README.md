# Parduotuve module

This module allows to buy/sell items in main shop.

## Example

```yaml
  # Buy max possible amount of items
  - _module: parduotuve
    action: pirkti
    item: MA1
    amount: 0
```
```yaml
  # Buy exactly 5 items
  - _module: parduotuve
    action: pirkti
    item: MA1
    amount: 5
```
```yaml
  # Buy exactly max-1 amount of items
  - _module: parduotuve
    action: pirkti
    item: MA1
    amount: -1
```
```yaml
  # Sell max amount of items
  - _module: parduotuve
    action: parduoti
    item: MA1
    amount: 0
```
