
enum ErrorCode {
  Success = 0,
  CommonInvalidParam1 = 100,
  CommonInvalidParam2 = 101,
  CommonInvalidParam3 = 102,
  CommonInvalidParam4 = 103,
  CommonInvalidParam5 = 104,
  CommonInvalidParam6 = 105,
  CommonInvalidParam7 = 106,
  CommonInvalidParam8 = 107,
  CommonInvalidParam9 = 108,
  CommonInvalidParam10 = 109,
  CommonInvalidParam11 = 110,
  CommonInvalidParam12 = 111,
  CommonInvalidState = 112,
  CommonInvalidStructure = 113,
  CommonIOError = 114,
  AnoncredsRevocationAccumulatorIsFull = 115,
  AnoncredsInvalidRevocationAccumulatorIndex = 116,
  AnoncredsCredentialRevoked = 117,
  AnoncredsProofRejected = 118,
};
typedef uintptr_t ErrorCode;
typedef ErrorCode (*FFITailTake)(const void *ctx, uint32_t idx, const void **tail_p);
typedef ErrorCode (*FFITailPut)(const void *ctx, const void *tail);

/**
 * Creates and returns credential definition (public and private keys, correctness proof) entities.
 *
 * Note that credential public key instances deallocation must be performed by
 * calling ursa_cl_credential_public_key_free.
 *
 * Note that credential private key instances deallocation must be performed by
 * calling ursa_cl_credential_private_key_free.
 *
 * Note that credential key correctness proof instances deallocation must be performed by
 * calling ursa_cl_credential_key_correctness_proof_free.
 *
 * # Arguments
 * * `credential_schema` - Reference that contains credential schema instance pointer.
 * * `non_credential_schema` - Reference that contains non credential schema instance pointer
 * * `support_revocation` - If true non revocation part of credential keys will be generated.
 * * `credential_pub_key_p` - Reference that will contain credential public key instance pointer.
 * * `credential_priv_key_p` - Reference that will contain credential private key instance pointer.
 * * `credential_key_correctness_proof_p` - Reference that will contain credential keys correctness proof instance pointer.
 */
ErrorCode ursa_cl_issuer_new_credential_def(const void *credential_schema,
                                            const void *non_credential_schema,
                                            _Bool support_revocation,
                                            const void **credential_pub_key_p,
                                            const void **credential_priv_key_p,
                                            const void **credential_key_correctness_proof_p);

/**
 * Adds new attribute to credential schema.
 *
 * # Arguments
 * * `credential_schema_builder` - Reference that contains credential schema builder instance pointer.
 * * `attr` - Attribute to add as null terminated string.
 */
ErrorCode ursa_cl_credential_schema_builder_add_attr(const void *credential_schema_builder,
                                                     const char *attr);

/**
 * Deallocates credential schema builder and returns credential schema entity instead.
 *
 * Note: Credentials schema instance deallocation must be performed by
 * calling ursa_cl_credential_schema_free.
 *
 * # Arguments
 * * `credential_schema_builder` - Reference that contains credential schema builder instance pointer
 * * `credential_schema_p` - Reference that will contain credentials schema instance pointer.
 */
ErrorCode ursa_cl_credential_schema_builder_finalize(const void *credential_schema_builder,
                                                     const void **credential_schema_p);

/**
 * Creates and returns credential schema entity builder.
 *
 * The purpose of credential schema builder is building of credential schema entity that
 * represents credential schema attributes set.
 *
 * Note: Credential schema builder instance deallocation must be performed by
 * calling ursa_cl_credential_schema_builder_finalize.
 *
 * # Arguments
 * * `credential_schema_builder_p` - Reference that will contain credentials attributes builder instance pointer.
 */
ErrorCode ursa_cl_credential_schema_builder_new(const void **credential_schema_builder_p);

/**
 * Deallocates credential schema instance.
 *
 * # Arguments
 * * `credential_schema` - Reference that contains credential schema instance pointer.
 */
ErrorCode ursa_cl_credential_schema_free(const void *credential_schema);

/**
 * Deallocates credential key correctness proof instance.
 *
 * # Arguments
 * * `credential_key_correctness_proof` - Reference that contains credential key correctness proof instance pointer.
 */
ErrorCode ursa_cl_credential_key_correctness_proof_free(const void *credential_key_correctness_proof);

/**
 * Creates and returns credential key correctness proof from json.
 *
 * Note: Credential key correctness proof instance deallocation must be performed
 * by calling ursa_cl_credential_key_correctness_proof_free
 *
 * # Arguments
 * * `credential_key_correctness_proof_json` - Reference that contains credential key correctness proof json.
 * * `credential_key_correctness_proof_p` - Reference that will contain credential key correctness proof instance pointer.
 */
ErrorCode ursa_cl_credential_key_correctness_proof_from_json(const char *credential_key_correctness_proof_json,
                                                             const void **credential_key_correctness_proof_p);

/**
 * Returns json representation of credential key correctness proof.
 *
 * # Arguments
 * * `credential_key_correctness_proof` - Reference that contains credential key correctness proof instance pointer.
 * * `credential_key_correctness_proof_p` - Reference that will contain credential key correctness proof json.
 */
ErrorCode ursa_cl_credential_key_correctness_proof_to_json(const void *credential_key_correctness_proof,
                                                           const char **credential_key_correctness_proof_json_p);

/**
 * Deallocates credential private key instance.
 *
 * # Arguments
 * * `credential_priv_key` - Reference that contains credential private key instance pointer.
 */
ErrorCode ursa_cl_credential_private_key_free(const void *credential_priv_key);

/**
 * Creates and returns credential private key from json.
 *
 * Note: Credential private key instance deallocation must be performed
 * by calling ursa_cl_credential_private_key_free
 *
 * # Arguments
 * * `credential_priv_key_json` - Reference that contains credential private key json.
 * * `credential_priv_key_p` - Reference that will contain credential private key instance pointer.
 */
ErrorCode ursa_cl_credential_private_key_from_json(const char *credential_priv_key_json,
                                                   const void **credential_priv_key_p);

/**
 * Returns json representation of credential private key.
 *
 * # Arguments
 * * `credential_priv_key` - Reference that contains credential private key instance pointer.
 * * `credential_pub_key_p` - Reference that will contain credential private key json.
 */
ErrorCode ursa_cl_credential_private_key_to_json(const void *credential_priv_key,
                                                 const char **credential_priv_key_json_p);

/**
 * Deallocates credential public key instance.
 *
 * # Arguments
 * * `credential_pub_key` - Reference that contains credential public key instance pointer.
 */
ErrorCode ursa_cl_credential_public_key_free(const void *credential_pub_key);

/**
 * Creates and returns credential public key from json.
 *
 * Note: Credential public key instance deallocation must be performed
 * by calling ursa_cl_credential_public_key_free
 *
 * # Arguments
 * * `credential_pub_key_json` - Reference that contains credential public key json.
 * * `credential_pub_key_p` - Reference that will contain credential public key instance pointer.
 */
ErrorCode ursa_cl_credential_public_key_from_json(const char *credential_pub_key_json,
                                                  const void **credential_pub_key_p);

/**
 * Returns json representation of credential public key.
 *
 * # Arguments
 * * `credential_pub_key` - Reference that contains credential public key instance pointer.
 * * `credential_pub_key_p` - Reference that will contain credential public key json.
 */
ErrorCode ursa_cl_credential_public_key_to_json(const void *credential_pub_key,
                                                const char **credential_pub_key_json_p);


/**
 * Adds new attribute to non credential schema.
 *
 * # Arguments
 * * `non_credential_schema_builder` - Reference that contains non credential schema builder instance pointer.
 * * `attr` - Attribute to add as null terminated string.
 */
ErrorCode ursa_cl_non_credential_schema_builder_add_attr(const void *non_credential_schema_builder,
                                                         const char *attr);

/**
 * Deallocates non_credential schema builder and returns non credential schema entity instead.
 *
 * Note: Non credential schema instance deallocation must be performed by
 * calling ursa_cl_non_credential_schema_free.
 *
 * # Arguments
 * * `non_credential_schema_builder` - Reference that contains non credential schema builder instance pointer
 * * `non_credential_schema_p` - Reference that will contain non credentials schema instance pointer.
 */
ErrorCode ursa_cl_non_credential_schema_builder_finalize(const void *non_credential_schema_builder,
                                                         const void **non_credential_schema_p);

/**
 * Creates and returns non credential schema builder.
 *
 * The purpose of non credential schema builder is building of non credential schema that
 * represents non credential schema attributes set. These are attributes added to schemas that are not on the ledger
 *
 * Note: Non credential schema builder instance deallocation must be performed by
 * calling ursa_cl_non_credential_schema_builder_finalize.
 *
 * # Arguments
 * * `credential_schema_builder_p` - Reference that will contain credentials attributes builder instance pointer.
 */
ErrorCode ursa_cl_non_credential_schema_builder_new(const void **non_credential_schema_builder_p);

/**
 * Deallocates credential schema instance.
 *
 * # Arguments
 * * `non_credential_schema` - Reference that contains non credential schema instance pointer.
 */
ErrorCode ursa_cl_non_credential_schema_free(const void *non_credential_schema);

/**
 * Adds new hidden attribute dec_value to credential values map.
 *
 * # Arguments
 * * `credential_values_builder` - Reference that contains credential values builder instance pointer.
 * * `attr` - Credential attr to add as null terminated string.
 * * `dec_value` - Credential attr dec_value. Decimal BigNum representation as null terminated string.
 * * `dec_blinding_factor` - Credential blinding factor. Decimal BigNum representation as null terminated string
 */
ErrorCode ursa_cl_credential_values_builder_add_dec_commitment(const void *credential_values_builder,
                                                               const char *attr,
                                                               const char *dec_value,
                                                               const char *dec_blinding_factor);

/**
 * Adds new hidden attribute dec_value to credential values map.
 *
 * # Arguments
 * * `credential_values_builder` - Reference that contains credential values builder instance pointer.
 * * `attr` - Credential attr to add as null terminated string.
 * * `dec_value` - Credential attr dec_value. Decimal BigNum representation as null terminated string.
 */
ErrorCode ursa_cl_credential_values_builder_add_dec_hidden(const void *credential_values_builder,
                                                           const char *attr,
                                                           const char *dec_value);

/**
 * Adds new known attribute dec_value to credential values map.
 *
 * # Arguments
 * * `credential_values_builder` - Reference that contains credential values builder instance pointer.
 * * `attr` - Credential attr to add as null terminated string.
 * * `dec_value` - Credential attr dec_value. Decimal BigNum representation as null terminated string.
 */
ErrorCode ursa_cl_credential_values_builder_add_dec_known(const void *credential_values_builder,
                                                          const char *attr,
                                                          const char *dec_value);

/**
 * Deallocates credential values builder and returns credential values entity instead.
 *
 * Note: Credentials values instance deallocation must be performed by
 * calling ursa_cl_credential_values_free.
 *
 * # Arguments
 * * `credential_values_builder` - Reference that contains credential attribute builder instance pointer.
 * * `credential_values_p` - Reference that will contain credentials values instance pointer.
 */
ErrorCode ursa_cl_credential_values_builder_finalize(const void *credential_values_builder,
                                                     const void **credential_values_p);

/**
 * Creates and returns credentials values entity builder.
 *
 * The purpose of credential values builder is building of credential values entity that
 * represents credential attributes values map.
 *
 * Note: Credentials values builder instance deallocation must be performed by
 * calling ursa_cl_credential_values_builder_finalize.
 *
 * # Arguments
 * * `credential_values_builder_p` - Reference that will contain credentials values builder instance pointer.
 */
ErrorCode ursa_cl_credential_values_builder_new(const void **credential_values_builder_p);

/**
 * Deallocates credential values instance.
 *
 * # Arguments
 * * `credential_values` - Credential values instance pointer
 */
ErrorCode ursa_cl_credential_values_free(const void *credential_values);

ErrorCode ursa_cl_issuer_merge_revocation_registry_deltas(const void *revoc_reg_delta,
                                                          const void *other_revoc_reg_delta,
                                                          const void **merged_revoc_reg_delta_p);

/**
 * Creates blinded credential secrets for given issuer key and master secret.
 *
 * Note that blinded credential secrets deallocation must be performed by
 * calling ursa_cl_blinded_credential_secrets_free.
 *
 * Note that credential secrets blinding factors deallocation must be performed by
 * calling ursa_cl_credential_secrets_blinding_factors_free.
 *
 * Note that blinded credential secrets correctness proof deallocation must be performed by
 * calling ursa_cl_blinded_credential_secrets_correctness_proof_free.
 *
 * # Arguments
 * * `credential_pub_key` - Reference that contains credential public key instance pointer.
 * * `credential_key_correctness_proof` - Reference that contains credential key correctness proof instance pointer.
 * * `credential_values` - Reference that contains credential values pointer.
 * * `credential_nonce` - Reference that contains nonce instance pointer.
 * * `blinded_credential_secrets_p` - Reference that will contain blinded credential secrets instance pointer.
 * * `credential_secrets_blinding_factors_p` - Reference that will contain credential secrets blinding factors instance pointer.
 * * `blinded_credential_secrets_correctness_proof_p` - Reference that will contain blinded credential secrets correctness proof instance pointer.
 */
ErrorCode ursa_cl_prover_blind_credential_secrets(const void *credential_pub_key,
                                                  const void *credential_key_correctness_proof,
                                                  const void *credential_values,
                                                  const void *credential_nonce,
                                                  const void **blinded_credential_secrets_p,
                                                  const void **credential_secrets_blinding_factors_p,
                                                  const void **blinded_credential_secrets_correctness_proof_p);


/**
 * Deallocates blinded credential secrets correctness proof instance.
 *
 * # Arguments
 * * `blinded_credential_secrets_correctness_proof` - Reference that contains blinded credential secrets correctness proof instance pointer.
 */
ErrorCode ursa_cl_blinded_credential_secrets_correctness_proof_free(const void *blinded_credential_secrets_correctness_proof);


/**
 * Returns json representation of blinded credential secrets correctness proof.
 *
 * # Arguments
 * * `blinded_credential_secrets_correctness_proof` - Reference that contains blinded credential secrets correctness proof pointer.
 * * `blinded_credential_secrets_correctness_proof_json_p` - Reference that will contain blinded credential secrets correctness proof json.
 */
ErrorCode ursa_cl_blinded_credential_secrets_correctness_proof_to_json(const void *blinded_credential_secrets_correctness_proof,
                                                                       const char **blinded_credential_secrets_correctness_proof_json_p);

/**
 * Deallocates  blinded credential secrets instance.
 *
 * # Arguments
 * * `blinded_credential_secrets` - Reference that contains blinded credential secrets instance pointer.
 */
ErrorCode ursa_cl_blinded_credential_secrets_free(const void *blinded_credential_secrets);


/**
 * Returns json representation of blinded credential secrets.
 *
 * # Arguments
 * * `blinded_credential_secrets` - Reference that contains Blinded credential secrets pointer.
 * * `blinded_credential_secrets_json_p` - Reference that will contain blinded credential secrets json.
 */
ErrorCode ursa_cl_blinded_credential_secrets_to_json(const void *blinded_credential_secrets,
                                                     const char **blinded_credential_secrets_json_p);

/**
 * Creates and returns blinded credential secrets from json.
 *
 * Note: Blinded credential secrets instance deallocation must be performed
 * by calling ursa_cl_blinded_credential_secrets_free
 *
 * # Arguments
 * * `blinded_credential_secrets_json` - Reference that contains blinded credential secret json.
 * * `blinded_credential_secrets_p` - Reference that will contain blinded credential secret instance pointer.
 */
ErrorCode ursa_cl_blinded_credential_secrets_from_json(const char *blinded_credential_secrets_json,
                                                       const void **blinded_credential_secrets_p);


/**
 * Creates and returns blinded credential secrets correctness proof json.
 *
 * Note: Blinded credential secrets correctness proof instance deallocation must be performed
 * by calling ursa_cl_blinded_credential_secrets_correctness_proof_free.
 *
 * # Arguments
 * * `blinded_credential_secrets_correctness_proof_json` - Reference that contains blinded credential secrets correctness proof json.
 * * `blinded_credential_secrets_correctness_proof_p` - Reference that will contain blinded credential secret correctness proof instance pointer.
 */
ErrorCode ursa_cl_blinded_credential_secrets_correctness_proof_from_json(const char *blinded_credential_secrets_correctness_proof_json,
                                                                         const void **blinded_credential_secrets_correctness_proof_p);

/**
 * Deallocates credential secrets blinding factors instance.
 *
 * # Arguments
 * * `credential_secrets_blinding_factors` - Reference that contains credential secrets blinding factors instance pointer.
 */
ErrorCode ursa_cl_credential_secrets_blinding_factors_free(const void *credential_secrets_blinding_factors);

/**
 * Creates and returns credential secrets blinding factors json.
 *
 * Note: Credential secrets blinding factors instance deallocation must be performed
 * by calling ursa_cl_credential_secrets_blinding_factors_free.
 *
 * # Arguments
 * * `credential_secrets_blinding_factors_json` - Reference that contains credential secrets blinding factors json.
 * * `credential_secrets_blinding_factors_p` - Reference that will contain credential secrets blinding factors instance pointer.
 */
ErrorCode ursa_cl_credential_secrets_blinding_factors_from_json(const char *credential_secrets_blinding_factors_json,
                                                                const void **credential_secrets_blinding_factors_p);

/**
 * Returns json representation of credential secrets blinding factors.
 *
 * # Arguments
 * * `credential_secrets_blinding_factors` - Reference that contains credential secrets blinding factors pointer.
 * * `credential_secrets_blinding_factors_json_p` - Reference that will contain credential secrets blinding factors json.
 */
ErrorCode ursa_cl_credential_secrets_blinding_factors_to_json(const void *credential_secrets_blinding_factors,
                                                              const char **credential_secrets_blinding_factors_json_p);


/**
 * Creates a master secret.
 *
 * Note that master secret deallocation must be performed by
 * calling ursa_cl_master_secret_free.
 *
 * # Arguments
 * * `master_secret_p` - Reference that will contain master secret instance pointer.
 */
ErrorCode ursa_cl_prover_new_master_secret(const void **master_secret_p);

/**
 * Deallocates master secret instance.
 *
 * # Arguments
 * * `master_secret` - Reference that contains master secret instance pointer.
 */
ErrorCode ursa_cl_master_secret_free(const void *master_secret);

/**
 * Creates and returns master secret from json.
 *
 * Note: Master secret instance deallocation must be performed
 * by calling ursa_cl_master_secret_free.
 *
 * # Arguments
 * * `master_secret_json` - Reference that contains master secret json.
 * * `master_secret_p` - Reference that will contain master secret instance pointer.
 */
ErrorCode ursa_cl_master_secret_from_json(const char *master_secret_json,
                                          const void **master_secret_p);

/**
 * Returns json representation of master secret.
 *
 * # Arguments
 * * `master_secret` - Reference that contains master secret instance pointer.
 * * `master_secret_json_p` - Reference that will contain master secret json.
 */
ErrorCode ursa_cl_master_secret_to_json(const void *master_secret,
                                        const char **master_secret_json_p);

/**
 * Signs credential values with primary keys only.
 *
 * Note that credential signature instances deallocation must be performed by
 * calling ursa_cl_credential_signature_free.
 *
 * Note that credential signature correctness proof instances deallocation must be performed by
 * calling ursa_cl_signature_correctness_proof_free.
 *
 * # Arguments
 * * `prover_id` - Prover identifier.
 * * `blinded_credential_secrets` - Blinded master secret instance pointer generated by Prover.
 * * `blinded_credential_secrets_correctness_proof` - Blinded master secret correctness proof instance pointer.
 * * `credential_nonce` - Nonce instance pointer used for verification of blinded_credential_secrets_correctness_proof.
 * * `credential_issuance_nonce` - Nonce instance pointer used for creation of signature_correctness_proof.
 * * `credential_values` - Credential values to be signed instance pointer.
 * * `credential_pub_key` - Credential public key instance pointer.
 * * `credential_priv_key` - Credential private key instance pointer.
 * * `credential_signature_p` - Reference that will contain credential signature instance pointer.
 * * `credential_signature_correctness_proof_p` - Reference that will contain credential signature correctness proof instance pointer.
 */
ErrorCode ursa_cl_issuer_sign_credential(const char *prover_id,
                                         const void *blinded_credential_secrets,
                                         const void *blinded_credential_secrets_correctness_proof,
                                         const void *credential_nonce,
                                         const void *credential_issuance_nonce,
                                         const void *credential_values,
                                         const void *credential_pub_key,
                                         const void *credential_priv_key,
                                         const void **credential_signature_p,
                                         const void **credential_signature_correctness_proof_p);

/**
 * Signs credential values with both primary and revocation keys.
 *
 *
 * Note that credential signature instances deallocation must be performed by
 * calling ursa_cl_credential_signature_free.
 *
 * Note that credential signature correctness proof instances deallocation must be performed by
 * calling ursa_cl_signature_correctness_proof_free.
 *
 *
 * Note that credential signature correctness proof instances deallocation must be performed by
 * calling ursa_cl_revocation_registry_delta_free.
 *
 * # Arguments
 * * `prover_id` - Prover identifier.
 * * `blinded_credential_secrets` - Blinded master secret instance pointer generated by Prover.
 * * `blinded_credential_secrets_correctness_proof` - Blinded master secret correctness proof instance pointer.
 * * `credential_nonce` - Nonce instance pointer used for verification of blinded_credential_secrets_correctness_proof.
 * * `credential_issuance_nonce` - Nonce instance pointer used for creation of signature_correctness_proof.
 * * `credential_values` - Credential values to be signed instance pointer.
 * * `credential_pub_key` - Credential public key instance pointer.
 * * `credential_priv_key` - Credential private key instance pointer.
 * * `rev_idx` - User index in revocation accumulator. Required for non-revocation credential_signature part generation.
 * * `max_cred_num` - Max credential number in generated registry.
 * * `rev_reg` - Revocation registry instance pointer.
 * * `rev_key_priv` - Revocation registry private key instance pointer.
 * * `credential_signature_p` - Reference that will contain credential signature instance pointer.
 * * `credential_signature_correctness_proof_p` - Reference that will contain credential signature correctness proof instance pointer.
 * * `revocation_registry_delta_p` - Reference that will contain revocation registry delta instance pointer.
 */
ErrorCode ursa_cl_issuer_sign_credential_with_revoc(const char *prover_id,
                                                    const void *blinded_credential_secrets,
                                                    const void *blinded_credential_secrets_correctness_proof,
                                                    const void *credential_nonce,
                                                    const void *credential_issuance_nonce,
                                                    const void *credential_values,
                                                    const void *credential_pub_key,
                                                    const void *credential_priv_key,
                                                    uint32_t rev_idx,
                                                    uint32_t max_cred_num,
                                                    _Bool issuance_by_default,
                                                    const void *rev_reg,
                                                    const void *rev_key_priv,
                                                    const void *ctx_tails,
                                                    FFITailTake take_tail,
                                                    FFITailPut put_tail,
                                                    const void **credential_signature_p,
                                                    const void **credential_signature_correctness_proof_p,
                                                    const void **revocation_registry_delta_p);


/**
 * Deallocates credential signature signature instance.
 *
 * # Arguments
 * * `credential_signature` - Reference that contains credential signature instance pointer.
 */
ErrorCode ursa_cl_credential_signature_free(const void *credential_signature);

/**
 * Creates and returns credential signature from json.
 *
 * Note: Credential signature instance deallocation must be performed
 * by calling ursa_cl_credential_signature_free
 *
 * # Arguments
 * * `credential_signature_json` - Reference that contains credential signature json.
 * * `credential_signature_p` - Reference that will contain credential signature instance pointer.
 */
ErrorCode ursa_cl_credential_signature_from_json(const char *credential_signature_json,
                                                 const void **credential_signature_p);

/**
 * Returns json representation of credential signature.
 *
 * # Arguments
 * * `credential_signature` - Reference that contains credential signature pointer.
 * * `credential_signature_json_p` - Reference that will contain credential signature json.
 */
ErrorCode ursa_cl_credential_signature_to_json(const void *credential_signature,
                                               const char **credential_signature_json_p);
/**
 * Get details for last occurred error.
 *
 * NOTE: Error is stored until the next one occurs.
 *       Returning pointer has the same lifetime.
 *
 * #Params
 * * `error_json_p` - Reference that will contain error details (if any error has occurred before)
 *  in the format:
 * {
 *     "backtrace": Optional<str> - error backtrace.
 *         Collecting of backtrace can be enabled by setting environment variable `RUST_BACKTRACE=1`
 *     "message": str - human-readable error description
 * }
 *
 */
void ursa_get_current_error(const char **error_json_p);


/**
 * Returns json representation of signature correctness proof.
 *
 * # Arguments
 * * `signature_correctness_proof` - Reference that contains signature correctness proof instance pointer.
 * * `signature_correctness_proof_json_p` - Reference that will contain signature correctness proof json.
 */
ErrorCode ursa_cl_signature_correctness_proof_to_json(const void *signature_correctness_proof,
                                                      const char **signature_correctness_proof_json_p);
