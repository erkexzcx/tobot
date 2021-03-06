# Kovojimas module

This module allows to perform actions in "Kovu laukas" with optional Slayer contracts.

# Examples

## Single hit killing

Simple example where you kill the mob with only 1 hit (indefinitely):

```yaml
  - _module: kovojimas
    vs: 0                # ID of enemy which to fight. Key in URL is 'vs'.
```

Same example, but you will only kill 100 enemies

```yaml
  - _module: kovojimas
    _count: 100
    vs: 0                # ID of enemy which to fight. Key in URL is 'vs'.
```

Also the same example, level up Slayer skill:

```yaml
  - _module: kovojimas
    vs: 0                # ID of enemy which to fight. Key in URL is 'vs'.
    slayer: 1            # ID of slayer contract (value is in the URL of "Žudyti 1-10 lygio karius")
```

## Multiple hits killing

If you don't kill the mob with 1 hit, you need to track your health and eat something to restore it:

```yaml
  - _module: eating    # This module ensures that you have 100% health before first hit
    food: UO10
  - _module: kovojimas
    vs: 99             # ID of enemy which to fight. Key in URL is 'vs'.
    food: UO10         # Food ID which to eat
    food_threshold: 50 # (OPTIONAL, default is 50 %) When health is below or equal to health level in %, continously eat to fully restore health
```

Same, but with slayer skill leveling up:

```yaml
  - _module: eating    # This module ensures that you have 100% health before first hit
    food: UO10
  - _module: kovojimas
    vs: 99             # ID of enemy which to fight. Key in URL is 'vs'.
    slayer: 10         # ID of slayer contract (value is in the URL of "Žudyti 1-10 lygio karius")
    food: UO10         # Food ID which to eat
    food_threshold: 50 # (OPTIONAL, default is 50 %) When health is below or equal to health level in %, continously eat to fully restore health
```

# _count usage

You only need `_count` field when enemy takes no health and you are not using `slayer` option.

Possible _completed unsuccessful_ endings (that lead to another task):

* Fight lost
* `slayer` contract completed
* Ran out of `food`.

# Slayer levels

When using bot you are likelly never going to see full list of slayer contracts without interrupting bot. Here it is :)

```
[*] Žudyti 1-10 lygio karius (min 0 slayer)
[*] Žudyti 11-30 lygio karius (min 20 slayer)
[*] Žudyti 31-50 lygio karius (min 50 slayer)
[*] Žudyti 51-90 lygio karius (min 100 slayer)
[*] Žudyti 91-150 lygio karius (min 300 slayer)
[*] Žudyti 151-200 lygio karius (min 600 slayer)
[*] Žudyti 201-260 lygio karius (min 1000 slayer)
[*] Žudyti 261-350 lygio karius (min 2000 slayer)
[*] Žudyti 351-470 lygio karius (min 3500 slayer)
[*] Žudyti 471-600 lygio karius (min 6000 slayer)
[*] Žudyti 601-750 lygio karius (min 10000 slayer)
[*] Žudyti 751-1000 lygio karius (min 25000 slayer)
[*] Žudyti 1001-1700 lygio karius (min 50000 slayer)
[*] Žudyti 1701-3000 lygio karius (min 100000 slayer)
[*] Žudyti 3001-8000 lygio karius (min 200000 slayer)
[*] Žudyti 8001-18000 lygio karius (min 500000 slayer)
```
