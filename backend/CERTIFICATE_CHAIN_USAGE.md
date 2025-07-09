# Certificate Chain Support in ORCA

This document explains how to use the certificate chain feature in ORCA for managing CyberArk instances that use certificates signed by intermediate Certificate Authorities (CAs).

## Overview

ORCA now supports uploading and managing complete certificate chains, which is essential when:
- Your CyberArk PVWA uses a certificate signed by an intermediate CA
- You have a multi-tier PKI infrastructure
- You need to trust custom root CAs not in the system trust store

## Key Features

1. **Automatic Chain Detection**: The system automatically identifies and categorizes certificates as root or intermediate
2. **Chain Validation**: Validates that intermediate certificates properly chain to root certificates
3. **Complete Chain Storage**: Stores all certificates in the chain for proper TLS validation
4. **Chain Information**: Provides detailed information about each certificate in the chain

## Usage Examples

### Uploading a Single Root CA

If your CyberArk instance uses a certificate directly signed by a root CA:

```bash
# Create a certificate authority with just the root CA
cat > root-ca.pem << EOF
-----BEGIN CERTIFICATE-----
[Your Root CA Certificate]
-----END CERTIFICATE-----
EOF

# Upload via API
curl -X POST https://orca.example.com/api/certificate-authorities \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Corporate Root CA",
    "description": "Main corporate root certificate authority",
    "certificate": "'$(cat root-ca.pem | jq -Rs .)'"
  }'
```

### Uploading a Certificate Chain

If your CyberArk instance uses a certificate signed by an intermediate CA:

```bash
# Create a certificate chain file (intermediate first, then root)
cat > ca-chain.pem << EOF
-----BEGIN CERTIFICATE-----
[Your Intermediate CA Certificate]
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
[Your Root CA Certificate]
-----END CERTIFICATE-----
EOF

# Upload via API
curl -X POST https://orca.example.com/api/certificate-authorities \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Corporate CA Chain",
    "description": "Root and intermediate CA chain for CyberArk",
    "certificate": "'$(cat ca-chain.pem | jq -Rs .)'"
  }'
```

## Certificate Chain Order

When uploading a certificate chain, the order matters:
1. **Start with the CA that signs your service certificates**: This is typically an intermediate CA
2. **Follow the chain upward**: Include each CA that signed the previous one
3. **End with the root CA**: The self-signed certificate should be last

Example chain order for a typical setup:
```
[Intermediate CA] - The CA that directly signs your CyberArk service certificate
[Root CA]         - The self-signed CA that signed the intermediate (last)
```

For a more complex chain:
```
[Intermediate CA 2] - Signs service certificates
[Intermediate CA 1] - Signs Intermediate CA 2
[Root CA]          - Signs Intermediate CA 1 (self-signed, last)
```

**Note**: You're uploading the CA certificates, not the service certificate itself. The service certificate remains on the CyberArk server.

## API Response Fields

When you upload a certificate chain, the API returns:

```json
{
  "id": "ca_01JZQB9ZBWTVCRV1D5RQAJV35T",
  "name": "Corporate CA Chain",
  "certificate_count": 2,
  "is_root_ca": false,        // True if primary cert is self-signed
  "is_intermediate": true,    // True if primary cert is intermediate
  "fingerprint": "f640e3c2...", // SHA256 of primary certificate
  "subject": "CN=Intermediate CA,O=Corp",
  "issuer": "CN=Root CA,O=Corp",
  "chain_info": [
    {
      "subject": "CN=Intermediate CA,O=Corp",
      "issuer": "CN=Root CA,O=Corp",
      "fingerprint": "f640e3c2...",
      "is_ca": true,
      "is_self_signed": false
    },
    {
      "subject": "CN=Root CA,O=Corp",
      "issuer": "CN=Root CA,O=Corp",
      "fingerprint": "b3c300a8...",
      "is_ca": true,
      "is_self_signed": true
    }
  ]
}
```

## Testing Your Configuration

After uploading your certificate chain:

1. **Test a CyberArk Connection**:
```bash
curl -X POST https://orca.example.com/api/cyberark/test-connection \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "base_url": "https://cyberark.example.com",
    "username": "test_user",
    "password": "test_password"
  }'
```

2. **Create a CyberArk Instance**:
```bash
curl -X POST https://orca.example.com/api/cyberark/instances \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production CyberArk",
    "base_url": "https://cyberark.example.com",
    "username": "api_user",
    "password": "api_password"
  }'
```

## Troubleshooting

### Common Issues

1. **"Certificate is not a CA certificate"**
   - Ensure you're uploading CA certificates, not service certificates
   - Check that the certificate has `CA:TRUE` in its basic constraints

2. **"Certificate chain validation failed"**
   - Verify the chain is complete (includes all intermediates up to root)
   - Check certificate validity dates
   - Ensure proper signing relationships

3. **Connection still fails after uploading CA**
   - Verify the service certificate is actually signed by your uploaded CA
   - Check that all intermediate certificates are included
   - Try setting `skip_tls_verify: true` temporarily to isolate TLS issues

### Viewing Certificate Details

To see details about uploaded certificate authorities:

```bash
# List all certificate authorities
curl -X GET https://orca.example.com/api/certificate-authorities \
  -H "Authorization: Bearer $TOKEN"

# Get specific certificate authority
curl -X GET https://orca.example.com/api/certificate-authorities/{id} \
  -H "Authorization: Bearer $TOKEN"
```

## Security Considerations

1. **Certificate Validation**: All uploaded certificates are validated to ensure they are proper CA certificates
2. **Chain Verification**: Certificate chains are verified to ensure proper signing relationships
3. **Automatic Updates**: When you modify certificate authorities, the certificate pool is automatically refreshed
4. **Isolation**: Each CyberArk instance uses the combined system and custom certificate pool

## Best Practices

1. **Upload Complete Chains**: Always include all certificates from the service certificate up to a trusted root
2. **Name Descriptively**: Use clear names that indicate the purpose and scope of the CA
3. **Document Chain Structure**: Use the description field to document the certificate hierarchy
4. **Regular Updates**: Update certificates before they expire to avoid service interruptions
5. **Test First**: Always test connections after uploading new certificates