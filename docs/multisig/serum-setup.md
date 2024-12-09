# Serum Multisig Setup

## Deploy Multisig Program:

1. Download the program binary code from the GH Release: [https://github.com/project-serum/multisig/releases/tag/v0.9.0](https://github.com/project-serum/multisig/releases/tag/v0.9.0)

```bash
# Create the program keypair used as multisig's `program-id`
solana-keygen new --outfile multisig-program.json

# Check the binary size in bytes
wc -c target/verifiable/serum_multisig.so
  234536 target/verifiable/serum_multisig.so

# Deploy the program, but add some buffer to the size as `max-len` (~2x)
solana program deploy \
  --keypair PATH_TO_KEYPAIR/id.json \ # optional, custom keypair
  --program-id multisig-program.json \
  --max-len 500000 \
  target/verifiable/serum_multisig.so
```

## Initialize Multisig State Acc (wallet):

```bash
# Use the create command from the gauntlet-serum-multisig package
yarn gauntlet serum_multisig:create --network=[NETWORK] --threshold=[THRESHOLD] [OWNERS...]
```

This will output 2 important addresses:

1.  Multisig State Acc Address: where the multisig program data for this instance is stored (e.g., threshold, owners, proposals).
    - Set this address as an env variable into `PATH_TO_GAUNTLET/networks/.env.network` as `MULTISIG_ADDRESS`
2.  Multisig Signer Address: address that will sign any transaction the multisig executes.
    - This is the address we need to transfer ownership to, as all proposals will be executed virtually by this signer.
    - This signer address is autogenerated from the Multisig Program ID and State Acc Address (can be derived/found at any time).