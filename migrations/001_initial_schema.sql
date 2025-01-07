-- migrations/001_initial_schema.sql
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create transactions table
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    description VARCHAR(50) NOT NULL,
    transaction_date DATE NOT NULL,
    amount_usd DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    CONSTRAINT transaction_description_length CHECK (LENGTH(description) <= 50),
    CONSTRAINT transaction_amount_positive CHECK (amount_usd > 0),
    CONSTRAINT transaction_date_not_future CHECK (transaction_date <= CURRENT_DATE),
    CONSTRAINT transaction_status_valid CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED'))
);

-- Create indexes for common queries
CREATE INDEX idx_transaction_date ON transactions(transaction_date);
CREATE INDEX idx_status ON transactions(status);
CREATE INDEX idx_created_at ON transactions(created_at);

-- Create audit log table for tracking changes
CREATE TABLE transaction_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    action VARCHAR(20) NOT NULL,
    old_status VARCHAR(20),
    new_status VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT audit_action_valid CHECK (action IN ('CREATE', 'UPDATE', 'STATUS_CHANGE'))
);

CREATE INDEX idx_audit_transaction_id ON transaction_audit_logs(transaction_id);
CREATE INDEX idx_audit_created_at ON transaction_audit_logs(created_at);

-- Function to update transaction audit log
CREATE OR REPLACE FUNCTION fn_log_transaction_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO transaction_audit_logs (transaction_id, action, new_status)
        VALUES (NEW.id, 'CREATE', NEW.status);
    ELSIF TG_OP = 'UPDATE' AND OLD.status != NEW.status THEN
        INSERT INTO transaction_audit_logs (transaction_id, action, old_status, new_status)
        VALUES (NEW.id, 'STATUS_CHANGE', OLD.status, NEW.status);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for transaction changes
CREATE TRIGGER trg_transaction_audit
AFTER INSERT OR UPDATE ON transactions
FOR EACH ROW
EXECUTE FUNCTION fn_log_transaction_changes();
