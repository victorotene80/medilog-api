# MediLog HTTP API Documentation

Base URL: `http://localhost:8080`

API prefix: `/api/v1`

## Conventions

JSON request bodies must use:

```http
Content-Type: application/json
```

Protected endpoints require:

```http
Authorization: Bearer <access_token>
```

Optional device metadata headers are supported by auth and rate-limit flows:

```http
X-Device-ID: ios-device-001
X-Device-Name: iPhone 15
X-Device-Fingerprint: demo-fingerprint
```

JSON responses use the shared envelope:

```json
{
  "status": true,
  "code": "MACHINE_READABLE_CODE",
  "message": "Human readable message",
  "data": {}
}
```

Error responses use the same envelope with `status: false`:

```json
{
  "status": false,
  "code": "VALIDATION_ERROR",
  "message": "One or more fields are invalid",
  "errors": "validation error details"
}
```

Common errors:

| HTTP status | Code | Meaning |
| --- | --- | --- |
| `400` | `INVALID_REQUEST_BODY` | Invalid JSON payload or unknown JSON fields. |
| `400` | `VALIDATION_ERROR` | Request fields failed validation. |
| `401` | `UNAUTHORIZED` or `UNAUTHENTICATED` | Token is missing, malformed, invalid, expired, or session context is missing. |
| `403` | `VERIFICATION_REQUIRED` | Account verification is required. |
| `403` | `ONBOARDING_REQUIRED` | Onboarding must be completed. |
| `403` | `ACCOUNT_NOT_ALLOWED` | Account is locked or blocked from access. |
| `429` | `RATE_LIMIT_EXCEEDED` | Too many requests. |
| `500` | `INTERNAL_SERVER_ERROR` | Unexpected server failure. |

## Route Access

Public endpoints do not require auth. Pre-onboarding endpoints require auth but do not require onboarding completion. Full app endpoints require both auth and completed onboarding.

| Method | Path | Access |
| --- | --- | --- |
| `GET` | `/health` | Public |
| `POST` | `/api/v1/auth/register` | Public |
| `POST` | `/api/v1/auth/login` | Public |
| `POST` | `/api/v1/auth/google` | Public |
| `POST` | `/api/v1/auth/forgot-password` | Public |
| `POST` | `/api/v1/auth/reset-password` | Public |
| `POST` | `/api/v1/auth/otp/request` | Public |
| `POST` | `/api/v1/auth/otp/verify` | Public |
| `POST` | `/api/v1/auth/otp/verify-onboarding` | Public |
| `POST` | `/api/v1/auth/logout` | Auth |
| `POST` | `/api/v1/auth/change-password` | Auth |
| `GET` | `/api/v1/reference/countries` | Public |
| `GET` | `/api/v1/reference/allergies/` | Public |
| `POST` | `/api/v1/reference/allergies/` | Auth |
| `PUT` | `/api/v1/reference/allergies/{id}` | Auth |
| `DELETE` | `/api/v1/reference/allergies/{id}` | Auth |
| `GET` | `/api/v1/reference/fun-facts/` | Public |
| `GET` | `/api/v1/reference/fun-facts/{id}` | Public |
| `POST` | `/api/v1/reference/fun-facts/` | Auth |
| `PUT` | `/api/v1/reference/fun-facts/{id}` | Auth |
| `DELETE` | `/api/v1/reference/fun-facts/{id}` | Auth |
| `GET` | `/api/v1/users/me` | Auth |
| `POST` | `/api/v1/emergency-contacts/` | Auth |
| `GET` | `/api/v1/health/allergies/` | Full app |
| `POST` | `/api/v1/health/allergies/` | Full app |
| `DELETE` | `/api/v1/health/allergies/{publicId}` | Full app |
| `GET` | `/api/v1/medications/` | Full app |
| `POST` | `/api/v1/medications/` | Full app |
| `GET` | `/api/v1/medications/{publicId}` | Full app |
| `PUT` | `/api/v1/medications/{publicId}` | Full app |
| `DELETE` | `/api/v1/medications/{publicId}` | Full app |
| `PATCH` | `/api/v1/medications/{publicId}/complete` | Full app |
| `POST` | `/api/v1/medications/{publicId}/adherence` | Full app |
| `GET` | `/api/v1/visits/` | Full app |
| `POST` | `/api/v1/visits/` | Full app |
| `GET` | `/api/v1/visits/{publicId}` | Full app |
| `PUT` | `/api/v1/visits/{publicId}` | Full app |
| `DELETE` | `/api/v1/visits/{publicId}` | Full app |
| `GET` | `/api/v1/ai/conversations/` | Full app |
| `POST` | `/api/v1/ai/conversations/` | Full app |
| `GET` | `/api/v1/ai/conversations/{publicId}` | Full app |
| `POST` | `/api/v1/ai/conversations/{publicId}/messages` | Full app |
| `PATCH` | `/api/v1/ai/conversations/{publicId}/archive` | Full app |
| `POST` | `/api/v1/drugs/verify` | Full app |
| `GET` | `/api/v1/drugs/scans` | Full app |
| `GET` | `/api/v1/drugs/scans/{publicId}` | Full app |
| `GET` | `/api/v1/dashboard` | Full app |

## Rate Limits

Rate limiting is configured for auth and OTP routes:

| Endpoint | Limit |
| --- | --- |
| `POST /api/v1/auth/register` | 5 per IP per hour, and 3 per IP plus email/phone per hour |
| `POST /api/v1/auth/login` | 5 per IP plus device plus email/phone per 10 minutes |
| `POST /api/v1/auth/google` | 10 per IP plus device per 10 minutes |
| `POST /api/v1/auth/forgot-password` | 3 per IP plus recipient per 10 minutes |
| `POST /api/v1/auth/reset-password` | 5 per IP plus recipient per 15 minutes |
| `POST /api/v1/auth/otp/request` | 50 per IP plus recipient plus purpose per 10 minutes |
| `POST /api/v1/auth/otp/verify` | 5 per IP plus recipient plus purpose per 10 minutes |
| `POST /api/v1/auth/otp/verify-onboarding` | 5 per IP plus device plus recipient plus purpose per 10 minutes |

## Health

### `GET /health`

Returns plain text when the server is alive.

Response: `200 OK`

```text
ok
```

```bash
curl -i "http://localhost:8080/health"
```

## Auth

### `POST /api/v1/auth/register`

Creates a new user account.

Request:

```json
{
  "first_name": "Victor",
  "last_name": "Otene",
  "email": "victor@example.com",
  "phone": "+2348012345678",
  "password": "Password123!",
  "dob": "1995-01-20",
  "sex": 1,
  "blood_type": "O+",
  "country_code": "NG"
}
```

Validation:

| Field | Notes |
| --- | --- |
| `first_name`, `last_name` | Required, 2 to 50 characters. |
| `email` | Optional, must be valid email when present. |
| `phone` | Optional, 8 to 15 characters when present. |
| `password` | Required, minimum 8 characters. |
| `dob` | Required string. |
| `sex` | Required integer. |
| `blood_type` | Required, max 3 characters. |
| `country_code` | Required, 1 to 4 characters. |

Success: `201 Created`, code `USER_CREATED`.

### `POST /api/v1/auth/login`

Logs in with either email or phone plus password.

Request:

```json
{
  "email": "victor@example.com",
  "password": "Password123!"
}
```

Alternative:

```json
{
  "phone": "+2348012345678",
  "password": "Password123!"
}
```

Success: `200 OK`, code `LOGIN_SUCCESS`.

Response data includes:

```json
{
  "user_id": "1",
  "tokens": {
    "access_token": "<access_token>",
    "access_token_expires_at": "2026-05-30T12:15:00Z",
    "refresh_token": "<refresh_token>",
    "refresh_token_expires_at": "2026-06-06T12:00:00Z"
  },
  "last_login": "2026-05-30T11:45:00Z",
  "challenge_id": null,
  "status": "SUCCESS",
  "onboarding_completed": false,
  "requires_onboarding": true
}
```

### `POST /api/v1/auth/google`

Authenticates with a Google ID token.

Request:

```json
{
  "idToken": "GOOGLE_ID_TOKEN_FROM_CLIENT"
}
```

Success: `200 OK`, code `GOOGLE_AUTH_SUCCESS`.

Existing-user response data may include tokens. New-user response data may include `status: "REGISTER"`, `google_id`, `picture_url`, and profile fields.

### `POST /api/v1/auth/forgot-password`

Requests a password reset OTP.

Request:

```json
{
  "recipient": "+2348012345678"
}
```

Success: `200 OK`, code `PASSWORD_RESET_OTP_SENT`.

### `POST /api/v1/auth/reset-password`

Resets password with a 6-digit OTP.

Request:

```json
{
  "recipient": "+2348012345678",
  "otp_code": "123456",
  "new_password": "NewPassword123!"
}
```

Success: `200 OK`, code `PASSWORD_RESET_SUCCESS`.

### `POST /api/v1/auth/otp/request`

Sends an OTP.

Request:

```json
{
  "recipient": "+2348012345678",
  "channel": "sms",
  "purpose": "phone_verification"
}
```

Allowed channels: `email`, `sms`, `whatsapp`.

Success: `200 OK`, code `OTP_SENT`.

### `POST /api/v1/auth/otp/verify`

Verifies a generic OTP.

Request:

```json
{
  "recipient": "+2348012345678",
  "code": "123456",
  "channel": "sms",
  "purpose": "phone_verification"
}
```

Code length: 4 to 8 characters.

Success: `200 OK`, code `OTP_VERIFIED`.

### `POST /api/v1/auth/otp/verify-onboarding`

Verifies onboarding OTP, activates the account, creates a session, and may return tokens.

Request:

```json
{
  "recipient": "+2348012345678",
  "code": "123456",
  "channel": "sms",
  "purpose": "phone_verification"
}
```

Success: `200 OK`, code `ONBOARDING_OTP_VERIFIED`.

Response data includes `verified`, optional `tokens`, `onboarding_completed`, `requires_onboarding`, and `onboarding_step`.

### `POST /api/v1/auth/logout`

Logs out the authenticated session.

Auth: yes

Success: `200 OK`, code `LOGOUT_SUCCESS`.

### `POST /api/v1/auth/change-password`

Changes the authenticated user's password.

Auth: yes

Request:

```json
{
  "old_password": "Password123!",
  "new_password": "NewPassword123!"
}
```

Success: `200 OK`, code `PASSWORD_CHANGED`.

## Reference Data

### `GET /api/v1/reference/countries`

Lists supported countries.

Success: `200 OK`, code `COUNTRIES_FETCHED`.

Response data item:

```json
{
  "code": "NG",
  "name": "Nigeria",
  "dial_code": "+234"
}
```

### `GET /api/v1/reference/allergies/`

Lists master allergy reference records.

Query:

| Name | Notes |
| --- | --- |
| `category` | Optional integer from 1 to 5. |

Success: `200 OK`, code `ALLERGIES_FETCHED`.

Response data item:

```json
{
  "id": "1",
  "name": "Peanuts",
  "category": 1,
  "category_label": "Food",
  "description": "Peanut allergy"
}
```

### `POST /api/v1/reference/allergies/`

Creates a master allergy reference record.

Auth: yes

Request:

```json
{
  "name": "Peanuts",
  "category": 1,
  "description": "Peanut allergy"
}
```

Validation: `name` required, max 120; `category` required, 1 to 5; `description` optional, max 500.

Success: `201 Created`, code `ALLERGY_CREATED`.

### `PUT /api/v1/reference/allergies/{id}`

Updates a master allergy reference record by numeric ID.

Auth: yes

Request body is the same as create.

Success: `200 OK`, code `ALLERGY_UPDATED`.

### `DELETE /api/v1/reference/allergies/{id}`

Deletes a master allergy reference record by numeric ID.

Auth: yes

Success: `200 OK`, code `ALLERGY_DELETED`.

### `GET /api/v1/reference/fun-facts/`

Lists fun facts.

Query:

| Name | Notes |
| --- | --- |
| `active_only` | Optional. Use `true` to return active facts only. |

Success: `200 OK`, code `FUN_FACTS_FETCHED`.

Response data item:

```json
{
  "id": "1",
  "title": "Hydration",
  "text": "Drinking water supports medication routines.",
  "category": "wellness",
  "target_country_code": "NG",
  "target_age_min": 18,
  "target_age_max": 65,
  "allergy_category": 1,
  "is_active": true,
  "created_at": "2026-05-30T10:00:00Z"
}
```

### `GET /api/v1/reference/fun-facts/{id}`

Gets one fun fact by numeric ID.

Success: `200 OK`, code `FUN_FACT_FETCHED`.

### `POST /api/v1/reference/fun-facts/`

Creates a fun fact.

Auth: yes

Request:

```json
{
  "title": "Hydration",
  "text": "Drinking water supports medication routines.",
  "category": "wellness",
  "target_country_code": "NG",
  "target_age_min": 18,
  "target_age_max": 65,
  "allergy_category": 1,
  "is_active": true
}
```

Validation: `text` required; `title` max 200; `category` max 100; `target_country_code` max 10; ages min 0; `allergy_category` 1 to 5.

Success: `201 Created`, code `FUN_FACT_CREATED`.

### `PUT /api/v1/reference/fun-facts/{id}`

Updates a fun fact by numeric ID.

Auth: yes

Request body is the same as create.

Success: `200 OK`, code `FUN_FACT_UPDATED`.

### `DELETE /api/v1/reference/fun-facts/{id}`

Deletes a fun fact by numeric ID.

Auth: yes

Success: `200 OK`, code `FUN_FACT_DELETED`.

## Users And Onboarding

### `GET /api/v1/users/me`

Returns the authenticated user profile. This endpoint is allowed before onboarding is completed.

Auth: yes

Success: `200 OK`, code `USER_FETCHED`.

Response data includes profile fields plus optional `emergency_contact`.

### `POST /api/v1/emergency-contacts/`

Creates an emergency contact for the authenticated user. This endpoint is allowed before onboarding is completed.

Auth: yes

Request:

```json
{
  "name": "Jane Otene",
  "relationship": "Sister",
  "phone": "+2348012345678",
  "country_code": "NG",
  "is_primary": true
}
```

Validation: `name` max 200; `relationship` max 100; `phone` must be E.164; `country_code` 1 to 4 characters.

Success: `201 Created`, code `CONTACT_CREATED`.

Current response data is an empty object.

## Health Allergies

Personal health allergy routes require completed onboarding.

### `GET /api/v1/health/allergies/`

Lists the authenticated user's allergy records.

Query:

| Name | Notes |
| --- | --- |
| `category` | Optional integer from 1 to 5. |

Success: `200 OK`, code `ALLERGIES_FETCHED`.

Response data item:

```json
{
  "id": "public-allergy-id",
  "allergy_id": 1,
  "name": "Peanuts",
  "description": "Causes swelling and rash",
  "severity": 4,
  "severity_label": "Life-threatening",
  "category": 1,
  "category_label": "Food",
  "is_custom": false,
  "created_at": "2026-05-30T10:00:00Z"
}
```

### `POST /api/v1/health/allergies/`

Records one or more allergies for the authenticated user.

Request:

```json
{
  "allergies": [
    {
      "allergy_id": 1,
      "name": "Peanuts",
      "category": 1,
      "severity": 4,
      "description": "Causes swelling and rash"
    }
  ]
}
```

Validation: `allergies` required with at least one item; `name` required max 120; `category` 1 to 5; `severity` optional 1 to 5; `description` optional max 500.

Success: `201 Created`, code `ALLERGIES_ADDED`.

### `DELETE /api/v1/health/allergies/{publicId}`

Deletes one allergy from the authenticated user's record by public ID.

Success: `200 OK`, code `ALLERGY_DELETED`.

## Medications

Medication routes require completed onboarding.

Medication request body:

```json
{
  "name": "Amoxicillin",
  "drug_class": 2,
  "dosage": "500mg",
  "frequency": "Twice daily",
  "with_food": true,
  "prescribed_by": "Dr Jane",
  "facility": "MediLog Clinic",
  "added_via": "manual",
  "registration_number": "A4-1234",
  "reg_country_code": "NG",
  "start_date": "2026-05-30T08:00:00Z",
  "end_date": "2026-06-06T08:00:00Z",
  "notes": "Take after meals",
  "times": [
    {
      "time_value": "08:00"
    },
    {
      "time_value": "20:00"
    }
  ]
}
```

Validation: `name` required max 200; `dosage` max 100; `frequency` max 100; `prescribed_by` max 150; `facility` max 200; `added_via` max 50; `registration_number` max 150; `reg_country_code` max 3; each `time_value` is required.

### `GET /api/v1/medications/`

Lists medications.

Query:

| Name | Notes |
| --- | --- |
| `active_only` | Optional. Use `true` to return active medications only. |

Success: `200 OK`, code `MEDICATIONS_FETCHED`.

### `POST /api/v1/medications/`

Creates a medication.

Success: `201 Created`, code `MEDICATION_CREATED`.

### `GET /api/v1/medications/{publicId}`

Gets one medication by public ID.

Success: `200 OK`, code `MEDICATION_FETCHED`.

Response data includes `public_id`, medication details, `is_verified`, `adherence_rate`, `is_completed`, `completed_date`, `times`, `created_at`, and `updated_at`.

### `PUT /api/v1/medications/{publicId}`

Updates one medication by public ID.

Request body is the same as create except `registration_number` and `reg_country_code` are not accepted by the update request.

Success: `200 OK`, code `MEDICATION_UPDATED`.

### `DELETE /api/v1/medications/{publicId}`

Deletes one medication by public ID.

Success: `200 OK`, code `MEDICATION_DELETED`.

### `PATCH /api/v1/medications/{publicId}/complete`

Marks one medication as completed.

Success: `200 OK`, code `MEDICATION_COMPLETED`.

### `POST /api/v1/medications/{publicId}/adherence`

Logs a medication dose.

Request:

```json
{
  "scheduled_at": "2026-05-30T08:00:00Z",
  "status": 1,
  "note": "Taken after breakfast"
}
```

Validation: `scheduled_at` required; `status` required, 1 to 3; `note` optional max 500.

Success: `201 Created`, code `ADHERENCE_LOGGED`.

## Visits

Visit routes require completed onboarding.

Visit request body:

```json
{
  "hospital_name": "MediLog Clinic",
  "diagnosis": "Malaria",
  "visit_date": "2026-05-30T09:00:00Z",
  "outcome": "Medication prescribed",
  "meds_count": 2,
  "doctor": "Dr Jane",
  "chief_complaint": "Fever",
  "notes": "Follow up in one week",
  "blood_pressure": "120/80",
  "temperature": 37.2,
  "weight": 75.5,
  "pulse": 78
}
```

Validation: `visit_date` required; `hospital_name` max 200; `diagnosis` max 500; `outcome` max 100; `meds_count` min 0; `doctor` max 150; `chief_complaint` max 500; `notes` max 1000; `blood_pressure` max 50; `temperature` 30 to 45; `weight` 1 to 500; `pulse` 30 to 250.

### `GET /api/v1/visits/`

Lists visits.

Query:

| Name | Notes |
| --- | --- |
| `from` | Optional RFC3339 datetime or `YYYY-MM-DD`. Must be paired with `to`. |
| `to` | Optional RFC3339 datetime or `YYYY-MM-DD`. Must be paired with `from`. |

Success: `200 OK`, code `VISITS_FETCHED`.

### `POST /api/v1/visits/`

Creates a visit.

Success: `201 Created`, code `VISIT_CREATED`.

### `GET /api/v1/visits/{publicId}`

Gets one visit by public ID.

Success: `200 OK`, code `VISIT_FETCHED`.

### `PUT /api/v1/visits/{publicId}`

Updates one visit by public ID.

Request body is the same as create.

Success: `200 OK`, code `VISIT_UPDATED`.

### `DELETE /api/v1/visits/{publicId}`

Deletes one visit by public ID.

Success: `200 OK`, code `VISIT_DELETED`.

## AI Conversations

AI routes require completed onboarding.

### `GET /api/v1/ai/conversations/`

Lists AI conversations.

Query:

| Name | Notes |
| --- | --- |
| `active_only` | Optional. Use `true` to return active conversations only. |

Success: `200 OK`, code `AI_CONVERSATIONS_FETCHED`.

### `POST /api/v1/ai/conversations/`

Creates an AI conversation.

Request:

```json
{
  "title": "Medication question",
  "related_medication_public_id": "med-public-id",
  "related_visit_public_id": "visit-public-id"
}
```

Validation: `title` optional max 200.

Success: `201 Created`, code `AI_CONVERSATION_CREATED`.

### `GET /api/v1/ai/conversations/{publicId}`

Gets one AI conversation by public ID.

Success: `200 OK`, code `AI_CONVERSATION_FETCHED`.

Response data includes `public_id`, `title`, optional related IDs, `summary`, `status`, `last_message_at`, optional `messages`, `created_at`, and `updated_at`.

### `POST /api/v1/ai/conversations/{publicId}/messages`

Sends a message and returns the user message plus AI reply.

Request:

```json
{
  "message": "Can I take this medication with food?",
  "language": "en"
}
```

Validation: `message` required, max 4000; `language` optional max 50.

Success: `201 Created`, code `AI_MESSAGE_SENT`.

Response data includes `conversation`, `message`, `reply`, `model`, `prompt_tokens`, `output_tokens`, and optional `context_meta`.

### `PATCH /api/v1/ai/conversations/{publicId}/archive`

Archives one AI conversation.

Success: `200 OK`, code `AI_CONVERSATION_ARCHIVED`.

## Drug Verification

Drug routes require completed onboarding.

### `POST /api/v1/drugs/verify`

Verifies a drug against the registry and records the scan.

Request:

```json
{
  "drug_name": "Amoxicillin",
  "registration_number": "A4-1234",
  "country_code": "NG",
  "expiry_date": "2027-12-31T00:00:00Z"
}
```

Validation: `drug_name` optional max 200; `registration_number` optional max 150; `country_code` required max 10; `expiry_date` optional.

Success: `200 OK`, code `DRUG_VERIFIED`.

Response data includes `public_id`, drug fields, verification status, explanation, confidence score, lot validity, optional registered medicine, and `created_at`.

### `GET /api/v1/drugs/scans`

Lists the authenticated user's drug scan history.

Success: `200 OK`, code `SCANS_FETCHED`.

### `GET /api/v1/drugs/scans/{publicId}`

Gets one drug scan by public ID.

Success: `200 OK`, code `SCAN_FETCHED`.

## Dashboard

### `GET /api/v1/dashboard`

Returns the authenticated user's dashboard.

Auth: full app

Success: `200 OK`, code `DASHBOARD_FETCHED`.

Response data shape:

```json
{
  "user": {
    "id": "1",
    "first_name": "Victor",
    "last_name": "Otene",
    "avatar_url": null
  },
  "medications": [],
  "health_overview": {
    "visits": {
      "total": 0,
      "upcoming": 0,
      "monthly": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]
    },
    "medications": {
      "completed": 0,
      "total": 0
    },
    "flagged_drugs": {
      "completed": 0,
      "total": 0
    }
  },
  "fun_fact": {
    "text": "A health tip appears here."
  }
}
```

## Curl Examples

Login:

```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: ios-device-001" \
  -H "X-Device-Name: iPhone 15" \
  -H "X-Device-Fingerprint: demo-fingerprint" \
  -d '{
    "email": "victor@example.com",
    "password": "Password123!"
  }'
```

List medications:

```bash
curl "http://localhost:8080/api/v1/medications/?active_only=true" \
  -H "Authorization: Bearer <access_token>"
```

Create a visit:

```bash
curl -X POST "http://localhost:8080/api/v1/visits/" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "hospital_name": "MediLog Clinic",
    "visit_date": "2026-05-30T09:00:00Z",
    "diagnosis": "Malaria"
  }'
```

Verify a drug:

```bash
curl -X POST "http://localhost:8080/api/v1/drugs/verify" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "drug_name": "Amoxicillin",
    "registration_number": "A4-1234",
    "country_code": "NG"
  }'
```
