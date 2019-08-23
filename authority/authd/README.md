# Use Vault To Build Own Certificate Authority

## Development

### Run Vault Dev Server Locally
`vault server -dev`

Export to `VAULT_ADDR` environment variable
`export VAULT_ADDR='http://127.0.0.1:8200'` **Make sure the address is correct. You should see it in the logging of the vault dev server**

Export `VAULT_TOKEN_ID` and `VAULT_TOKEN` environment variables.
You should see them in the vault dev log output. 

**NOTE: The Vault dev server stores everyting in memory. So if it crashes or is stopped manually you have to reset these values, as they will likely change on the next startup**

### Add the proper policy

To execute below CLI commands we need the minimum policy within [ca_policy.hcl](vault/ca_policy.hcl)
Enter: `vault policy write ca_policy vault/ca_policy.hcl`
To make sure if worked list out all policies using `vault policy list`. You should see it in there.

### Generate the root CA

1. Enable the secrets pki engine
```
$ vault secrets enable pki
```

2. Tune the pki secrets engine to issue certificates with a maximum time-to-live (TTL) of 87600 hours.
```
$ vault secrets tune -max-lease-ttl=87600h pki
```

3. Generate the root certificate and save it to `certs/CA_cert.crt`
```
$ vault write -field=certificate pki/root/generate/internal \
        common_name="<YOUR_COMMON_NAME>" \
        ttl=87600h > CA_cert.crt
```

4. Configure the CA and CRL URLs:
```
$ vault write pki/config/urls \
        issuing_certificates="http://127.0.0.1:8200/v1/pki/ca" \
        crl_distribution_points="http://127.0.0.1:8200/v1/pki/crl"
```

### Generate Intermediate CA

1. First, enable the pki secrets engine at the pki_int path
```
$ vault secrets enable -path=pki_int pki
```

2. Tune the pki_int secrets engine to issue certificates with a maximum time-to-live (TTL) of 43800 hours.
```
$ vault secrets tune -max-lease-ttl=43800h pki_int
```

3. Execute the following command to generate an intermediate and save the CSR as pki_intermediate.csr:
```
$ vault write -format=json pki_int/intermediate/generate/internal \
        common_name="<YOUR_COMMON_NAME> Intermediate Authority" ttl="43800h" \
        | jq -r '.data.csr' > pki_intermediate.csr
```

4. Sign the intermediate certificate with the root certificate and save the generated certificate as intermediate.cert.pem:
```
$ vault write -format=json pki/root/sign-intermediate csr=@pki_intermediate.csr \
        format=pem_bundle ttl="43800h" \
        | jq -r '.data.certificate' > intermediate.cert.pem
```

5. Once the CSR is signed and the root CA returns a certificate, it can be imported back into Vault:
```
$ vault write pki_int/intermediate/set-signed certificate=@intermediate.cert.pem
```


### Create a Role
We create a role that is allowed to issue certificates for YOUR_COMMON_NAME and allow subdomains
```
$ vault write pki_int/roles/<ROLE_NAME> \
        allowed_domains="<YOUR_COMMON_NAME>" \
        allow_subdomains=true \
        max_ttl="720h"
```

### Request Certificates
```
$ vault write pki_int/issue/example-dot-com common_name="<subdomain>.<your_common_name>" ttl="24h"
```

Or via curl
```
$ curl --header "X-Vault-Token: <YOUR_VAULT_TOKEN>" \
       --request POST \
       --data '{"common_name": "<subdomain>.<your_common_name>", "ttl": "24h"}' \
       https://127.0.0.1:8200/v1/pki_int/issue/example-dot-com | jq
```






