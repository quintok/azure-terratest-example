# azure-terratest-example
Example showing a CI pipeline for terratest

The idea here is we use terraform to provision some environment.  This then uses terratest to confirm the environment is in a fit state.

Ideally this would be part of the testing framework for a module to test different permutations of how it could work.

# Future Goals
- Allow feature branch testing
- Promotion potentially between environments
