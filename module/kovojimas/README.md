# Kovojimas module

This module allows to perform actions in "Kovu laukas" with optional Slayer contracts.

## Example - 1 hit mob killing

Simple example where you kill the mob with only 1 hit:

```yaml
  - _module: kovojimas
    vs: 0                # ID of enemy which to fight. Key in URL is 'vs'.
```

Also the same example, level up Slayer skill:

```yaml
  - _module: kovojimas
    vs: 0                # ID of enemy which to fight. Key in URL is 'vs'.
    slayer: 1            # ID of slayer contract (value is in the URL of "Žudyti 1-10 lygio karius")
```

## Example - mob killing with multiple hits

If you don't kill the mob with 1 hit, you need to track your health and eat something to restore it:

```yaml
  - _module: eating      # This module ensures that you have 100% health before first hit
    food: UO10
  - _module: kovojimas
    vs: 99               # ID of enemy which to fight. Key in URL is 'vs'.
    eating: UO10         # Food ID which to eat
    eating_threshold: 50 # (OPTIONAL, default is 50 %) When health is below or equal to health level in %, continously eat to fully restore health
```

Same, but with slayer skill leveling up:

```yaml
  - _module: eating      # This module ensures that you have 100% health before first hit
    food: UO10
  - _module: kovojimas
    vs: 99               # ID of enemy which to fight. Key in URL is 'vs'.
    slayer: 10           # ID of slayer contract (value is in the URL of "Žudyti 1-10 lygio karius")
    eating: UO10         # Food ID which to eat
    eating_threshold: 50 # (OPTIONAL, default is 50 %) When health is below or equal to health level in %, continously eat to fully restore health
```
