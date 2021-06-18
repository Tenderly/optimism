# @eth-optimism/integration-tests

## Setup

Follow installation + build instructions in the [primary README](../README.md).
Then, run:

```
yarn build:integration
```

## Running tests

### Testing a production network

Create an `.env` file and fill it out.
Look at `.env.example` to know which variables to include.

Once you have your environment set up, run:
```bash
IS_PROD_NETWORK=true yarn test:integration
```

You MUST set `IS_PROD_NETWORK=true` on the command line or nothing will work.
You can also set environment variables on the command line instead of inside `.env` if you want:
```bash
L1_URL=whatever L2_URL=whatever IS_PROD_NETWORK=true yarn test:integration
```

Note that this can take an extremely long time (~1hr).
