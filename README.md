# Op-Sep: Separate Your Data from your Keys
Use strong cryptography and simple rate-limting to operationally separate your data from your keys.
Reduce the damage of data-breaches 100x (and maintain an audit log to measure exposure).

## Setup

### Fetch repo:
```bash
$ git clone git@github.com:opsep/opsep-server.git && cd opsep-server
```

### Run the server
Use [reflex](https://github.com/cespare/reflex) to reload the server in development:
```bash
$ ./run_localhost.sh
```
(install via `$ go get github.com/cespare/reflex`)

### Test that it's working:
```bash
$ curl localhost:8080/ping
```

## Use

### Encryption

Encrypt a key locally (this would be randomly generated by your client library):
```bash
$ to_decrypt=$(echo "{\"key\":\"00000000000000000000000000000000\"}" | openssl pkeyutl -encrypt -pubin -inkey insecure_certs/insecure_crt.pub -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:sha256 -pkeyopt rsa_mgf1_md:sha256 | base64)
```
(you can see this with `$ echo $to_decrypt`)

### Decryption

Make an API call to decrypt the file you just made:
```bash
$ curl -X POST localhost:8080/api/v1/decrypt -H 'Content-Type: application/json' -d '{"key_retrieval_ciphertext":"'$(echo $to_decrypt)'"}'
{"key_recovered":"00000000000000000000000000000000","request_sha256":"3632...e6f0","ratelimit_limit":100,"ratelimit_remaining":94,"ratelimit_resets_in":197}
```

### Key Deprecation
Make your data auto-expire at a certain date (not that the data can still be recovered if you have your Key Encryption Key, but it is not accessible via this service:
```bash
$ curl -X POST localhost:8080/api/v1/decrypt -H 'Content-Type: application/json' -d '{"key_retrieval_ciphertext":"'$(echo "{\"key\":\"00000000000000000000000000000000\", \"deprecate_at\":\"2020-01-01T12:00:00Z\"}" | openssl pkeyutl -encrypt -pubin -inkey insecure_certs/insecure_crt.pub -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:sha256 -pkeyopt rsa_mgf1_md:sha256 | base64)'"}'
{"error_name":"DeprecatedDecryptionKeyError","error_description":"Key to decrypt this payload marked as deprecated."}
```

### Audit Logging:
Calculate the hash of the file your decryption request:
```bash
$ echo $to_decrypt | base64 --decode | shasum -a 256
3632dc12b3b03c4508ce7155941f249a2ec521c000619a345a7a186f7fa9e6f0  -
```
(your result will be different, as asymmetric encryption uses a randomly generated nonce each time it is run)

Query to see all past decrypts for a specific record:
```bash
$ curl localhost:8080/api/v1/logs/3632dc12b3b03c4508ce7155941f249a2ec521c000619a345a7a186f7fa9e6f0
[{"id":2,"created_at":"2020-08-10T23:14:40Z","request_ip_address":"127.0.0.1","request_user_agent":"curl/7.64.1"},{"id":3,"created_at":"2020-08-10T23:14:42Z","request_ip_address":"127.0.0.1","request_user_agent":"curl/7.64.1"}]
```
(this will look pretty-print if you append `| python -m json.tool`. Power users can check out [jq](https://stedolan.github.io/jq/)
)

Worried about a breach? See all decrypts as CSV:
```bash
$ sqlite3 opsep.sqlite3 -header -csv 'select * from api_calls;'
```

### Details

#### One-Liner
Test in one-line:
```bash
$ curl -X POST localhost:8080/api/v1/decrypt -H 'Content-Type: application/json' -d '{"key_retrieval_ciphertext":"'$(echo "{\"key\":\"00000000000000000000000000000000\"}" | openssl pkeyutl -encrypt -pubin -inkey insecure_certs/insecure_crt.pub -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:sha256 -pkeyopt rsa_mgf1_md:sha256 | base64)'"}'
{"key_recovered":"00000000000000000000000000000000","request_sha256":"3632...e6f0","ratelimit_limit":100,"ratelimit_remaining":93,"ratelimit_resets_in":122}
```

#### Rate-Limit
If you want a rough test of 429-ing, you can do this:
```bash
$ for i in {1..99}; do curl [...] "http://localhost:8080/api/v1/decrypt" ; done
```

#### Create an (insecure) RSA keypair
```bash
$ openssl genrsa -out insecure_pem.priv 4096 && openssl rsa -in insecure_pem.priv -pubout -out insecure_crt.pub
Generating RSA private key, 4096 bit long modulus (2 primes)
...........................................++++
...............................++++
e is 65537 (0x010001)
writing RSA key
```

Query decryption API call logs:
```bash
$ curl https://www.secondguard.com/callz/100/0 | jq | grep String | uniq
```
