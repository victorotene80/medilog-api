-- MediLog API Initial Schema
-- Generated from GORM models

-- ============================================================
-- USERS
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid(),
    email TEXT,
    phone TEXT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    avatar_url TEXT,
    date_of_birth TIMESTAMPTZ,
    sex INTEGER,
    blood_type TEXT,
    country_code TEXT,
    password_hash TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    status TEXT NOT NULL DEFAULT 'active',
    email_verified_at TIMESTAMPTZ,
    phone_verified_at TIMESTAMPTZ,
    is_onboarding_completed BOOLEAN NOT NULL DEFAULT false,
    password_changed_at TIMESTAMPTZ,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    last_login_ip TEXT,
    last_active_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- USER PROFILES
-- ============================================================
CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    weight DOUBLE PRECISION,
    height DOUBLE PRECISION,
    weight_unit TEXT NOT NULL DEFAULT 'kg',
    temperature_unit TEXT NOT NULL DEFAULT 'celsius',
    ai_questions_used INTEGER NOT NULL DEFAULT 0,
    ai_questions_total INTEGER NOT NULL DEFAULT 10,
    ai_is_pro BOOLEAN NOT NULL DEFAULT false,
    ai_questions_reset_at TIMESTAMPTZ,
    medication_reminders_enabled BOOLEAN NOT NULL DEFAULT true,
    refill_reminders_enabled BOOLEAN NOT NULL DEFAULT true,
    appointment_reminders_enabled BOOLEAN NOT NULL DEFAULT true,
    ai_health_tips_enabled BOOLEAN NOT NULL DEFAULT true,
    support_updates_enabled BOOLEAN NOT NULL DEFAULT true,
    app_updates_enabled BOOLEAN NOT NULL DEFAULT true,
    push_enabled BOOLEAN NOT NULL DEFAULT true,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    sms_enabled BOOLEAN NOT NULL DEFAULT false,
    whatsapp_enabled BOOLEAN NOT NULL DEFAULT false
);

-- ============================================================
-- USER AUTH PROVIDERS
-- ============================================================
CREATE TABLE IF NOT EXISTS user_auth_providers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    provider TEXT NOT NULL,
    provider_uid TEXT NOT NULL,
    email TEXT,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_auth_providers_user_id ON user_auth_providers(user_id);

-- ============================================================
-- REFRESH TOKENS
-- ============================================================
CREATE TABLE IF NOT EXISTS refresh_token (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token_hash TEXT NOT NULL,
    device_id TEXT,
    device_name TEXT,
    ip_address TEXT,
    user_agent TEXT,
    device_fingerprint VARCHAR(150),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    replaced_by_token_hash TEXT,
    date_created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_refresh_token_user_id ON refresh_token(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_token_deleted_at ON refresh_token(deleted_at);

-- ============================================================
-- OTP CODES
-- ============================================================
CREATE TABLE IF NOT EXISTS otp_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    recipient TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    channel TEXT NOT NULL,
    purpose TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_otp_codes_user_id ON otp_codes(user_id);

-- ============================================================
-- MEDICATIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS medications (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    drug_class INTEGER,
    dosage TEXT,
    frequency TEXT,
    with_food BOOLEAN NOT NULL DEFAULT false,
    prescribed_by TEXT,
    facility TEXT,
    added_via TEXT,
    registration_number TEXT,
    reg_country_code TEXT,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    notes TEXT,
    adherence_count INTEGER NOT NULL DEFAULT 0,
    total_doses INTEGER NOT NULL DEFAULT 0,
    is_completed BOOLEAN NOT NULL DEFAULT false,
    completed_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_medications_user_id ON medications(user_id);
CREATE INDEX IF NOT EXISTS idx_medications_deleted_at ON medications(deleted_at);

-- ============================================================
-- MEDICATION TIMES
-- ============================================================
CREATE TABLE IF NOT EXISTS medication_times (
    id BIGSERIAL PRIMARY KEY,
    medication_id BIGINT NOT NULL,
    time_value TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_medication_times_medication_id ON medication_times(medication_id);
CREATE INDEX IF NOT EXISTS idx_medication_times_deleted_at ON medication_times(deleted_at);

-- ============================================================
-- MEDICATION ADHERENCE LOGS
-- ============================================================
CREATE TABLE IF NOT EXISTS medication_adherence_logs (
    id BIGSERIAL PRIMARY KEY,
    medication_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    taken_at TIMESTAMPTZ,
    status INTEGER NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_medication_adherence_logs_deleted_at ON medication_adherence_logs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_medication_adherence_logs_scheduled_at ON medication_adherence_logs(scheduled_at);

-- ============================================================
-- VISITS
-- ============================================================
CREATE TABLE IF NOT EXISTS visits (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    hospital_name TEXT,
    diagnosis TEXT,
    visit_date TIMESTAMPTZ NOT NULL,
    outcome TEXT,
    meds_count INTEGER NOT NULL DEFAULT 0,
    doctor TEXT,
    chief_complaint TEXT,
    notes TEXT,
    blood_pressure TEXT,
    temperature DOUBLE PRECISION,
    weight DOUBLE PRECISION,
    pulse INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_visits_user_id ON visits(user_id);
CREATE INDEX IF NOT EXISTS idx_visits_visit_date ON visits(visit_date);
CREATE INDEX IF NOT EXISTS idx_visits_deleted_at ON visits(deleted_at);

-- ============================================================
-- ALLERGIES
-- ============================================================
CREATE TABLE IF NOT EXISTS allergies (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category INTEGER NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_allergies_deleted_at ON allergies(deleted_at);

-- ============================================================
-- USER ALLERGIES
-- ============================================================
CREATE TABLE IF NOT EXISTS user_allergies (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    allergy_id BIGINT,
    name TEXT NOT NULL,
    description TEXT,
    severity SMALLINT,
    category INTEGER NOT NULL,
    created_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_user_allergies_deleted_at ON user_allergies(deleted_at);

-- ============================================================
-- EMERGENCY CONTACTS
-- ============================================================
CREATE TABLE IF NOT EXISTS emergency_contacts (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    relationship TEXT NOT NULL,
    phone TEXT NOT NULL,
    country_code TEXT NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_emergency_contacts_deleted_at ON emergency_contacts(deleted_at);

-- ============================================================
-- DRUG SCANS
-- ============================================================
CREATE TABLE IF NOT EXISTS drug_scans (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    drug_name TEXT,
    registration_number TEXT,
    expiry_date TIMESTAMPTZ,
    regulatory_body_id BIGINT,
    registered_medicine_id BIGINT,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    confidence_score DOUBLE PRECISION,
    lot_number_valid BOOLEAN,
    verification_status TEXT,
    explanation TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_drug_scans_user_id ON drug_scans(user_id);
CREATE INDEX IF NOT EXISTS idx_drug_scans_deleted_at ON drug_scans(deleted_at);

-- ============================================================
-- REGISTERED MEDICINES
-- ============================================================
CREATE TABLE IF NOT EXISTS registered_medicines (
    id BIGSERIAL PRIMARY KEY,
    regulatory_body_id BIGINT NOT NULL,
    country_code TEXT NOT NULL,
    drug_name TEXT NOT NULL,
    registration_number TEXT NOT NULL,
    barcode TEXT,
    manufacturer TEXT,
    registered_date TIMESTAMPTZ,
    expiry_date TIMESTAMPTZ,
    status BOOLEAN NOT NULL,
    source_product_id BIGINT,
    strength TEXT,
    ingredient_name TEXT,
    synonym TEXT,
    category_name TEXT,
    form_name TEXT,
    route_name TEXT,
    applicant_name TEXT,
    source_payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_registered_medicines_deleted_at ON registered_medicines(deleted_at);

-- ============================================================
-- COUNTRIES
-- ============================================================
CREATE TABLE IF NOT EXISTS countries (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    dial_code TEXT NOT NULL
);

-- ============================================================
-- FUN FACTS
-- ============================================================
CREATE TABLE IF NOT EXISTS fun_facts (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    text TEXT NOT NULL,
    category TEXT,
    target_country_code TEXT,
    target_age_min INTEGER,
    target_age_max INTEGER,
    allergy_category INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_fun_facts_deleted_at ON fun_facts(deleted_at);

-- ============================================================
-- AI CONVERSATIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_conversations (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    title TEXT,
    related_medication_id BIGINT,
    related_visit_id BIGINT,
    summary TEXT,
    summary_up_to BIGINT,
    status TEXT NOT NULL DEFAULT 'active',
    last_message_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_user_id ON ai_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_deleted_at ON ai_conversations(deleted_at);

-- ============================================================
-- AI MESSAGES
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    token_count INTEGER,
    meta JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_messages_conversation_created ON ai_messages(conversation_id, created_at);

-- ============================================================
-- SUPPORT TICKETS
-- ============================================================
CREATE TABLE IF NOT EXISTS support_tickets (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    category_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    priority TEXT NOT NULL DEFAULT 'normal',
    assigned_to BIGINT,
    last_message_at TIMESTAMPTZ,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    date_assigned TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_support_tickets_user_id ON support_tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_status ON support_tickets(status);

-- ============================================================
-- SUPPORT MESSAGES
-- ============================================================
CREATE TABLE IF NOT EXISTS support_messages (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL,
    sender_user_id BIGINT,
    message TEXT NOT NULL,
    is_internal_note BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_support_messages_ticket_id ON support_messages(ticket_id);

-- ============================================================
-- SUPPORT ATTACHMENTS
-- ============================================================
CREATE TABLE IF NOT EXISTS support_attachments (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL,
    file_url TEXT NOT NULL,
    file_name TEXT,
    file_type TEXT,
    file_size BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_support_attachments_message_id ON support_attachments(message_id);

-- ============================================================
-- FEEDBACK
-- ============================================================
CREATE TABLE IF NOT EXISTS feedback (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid(),
    user_id BIGINT,
    rating INTEGER,
    title TEXT,
    message TEXT NOT NULL,
    app_version TEXT,
    platform TEXT,
    device_model TEXT,
    status TEXT NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- NOTIFICATIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    type TEXT NOT NULL,
    channel TEXT,
    status TEXT NOT NULL DEFAULT 'unread',
    image_url TEXT,
    scheduled_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);

-- ============================================================
-- OUTBOX EVENTS
-- ============================================================
CREATE TABLE IF NOT EXISTS outbox_events (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    kind VARCHAR(50) NOT NULL,
    aggregate_id BIGINT NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL,
    metadata JSONB NOT NULL,
    correlation_id VARCHAR(100),
    causation_id VARCHAR(100),
    version INTEGER NOT NULL,
    status INTEGER NOT NULL,
    attempts INTEGER NOT NULL,
    last_error TEXT,
    in_progress_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_outbox_events_aggregate_id ON outbox_events(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_outbox_events_occurred_at ON outbox_events(occurred_at);
CREATE INDEX IF NOT EXISTS idx_outbox_events_status ON outbox_events(status);

-- ============================================================
-- AUDIT LOGS
-- ============================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    action TEXT NOT NULL,
    user_id TEXT,
    actor_id TEXT,
    session_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    country_code TEXT,
    target_id TEXT,
    metadata JSONB,
    success BOOLEAN NOT NULL DEFAULT true,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
