-- Fix Vendor Profiles for Testing
-- Creates vendor profiles for existing users who need vendor access

-- ====================================================================
-- CREATE VENDOR PROFILE FOR USER ID 2
-- ====================================================================

-- First, check if user ID 2 exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE id = 2) THEN
        -- Update user role to vendor if needed
        UPDATE users
        SET role = 'vendor'
        WHERE id = 2 AND role != 'vendor';

        -- Create vendor profile if it doesn't exist
        INSERT INTO vendors (
            user_id,
            company_name,
            city,
            verified,
            created_at,
            updated_at
        )
        SELECT
            2,
            'Test Vendor Business',
            'San Francisco',
            true,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        WHERE NOT EXISTS (
            SELECT 1 FROM vendors WHERE user_id = 2
        );

        RAISE NOTICE 'Vendor profile created for user ID 2';
    ELSE
        RAISE NOTICE 'User ID 2 does not exist, skipping vendor profile creation';
    END IF;
END $$;

-- ====================================================================
-- CREATE VENDOR PROFILE FOR USER ID 3 (if exists)
-- ====================================================================

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE id = 3) THEN
        UPDATE users
        SET role = 'vendor'
        WHERE id = 3 AND role != 'vendor';

        INSERT INTO vendors (
            user_id,
            company_name,
            city,
            verified,
            created_at,
            updated_at
        )
        SELECT
            3,
            'Second Test Vendor',
            'San Francisco',
            true,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        WHERE NOT EXISTS (
            SELECT 1 FROM vendors WHERE user_id = 3
        );

        RAISE NOTICE 'Vendor profile created for user ID 3';
    END IF;
END $$;

-- ====================================================================
-- VERIFY CREATION
-- ====================================================================

-- Show created vendor profiles
SELECT
    v.id,
    v.user_id,
    u.name as user_name,
    u.email as user_email,
    u.role,
    v.company_name,
    v.city,
    v.verified
FROM vendors v
LEFT JOIN users u ON v.user_id = u.id
WHERE v.user_id IN (2, 3)
ORDER BY v.user_id;
