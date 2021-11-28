# Trolis module

This module allows to fight trolis ("Miestas --> Trolis").

## Examples

Without eating (assuming you have full health and nice stats so your current health will be enough until troll dies):

```yaml
  - _module: trolis
```

More advanced, with eating:

```yaml
  - _module: eating
    food: UO1
  - _module: trolis
    food: UO1
```

**NOTE**: There is no implementation of waiting until Trolis appears. It just goes to the next task, which might lead to page refresh-loop if the only task in activity is `trolis`...
