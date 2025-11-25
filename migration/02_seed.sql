INSERT INTO teams (team_name) VALUES
    ('backend'),
    ('frontend'),
    ('devops'),
    ('mobile')
ON CONFLICT (team_name) DO NOTHING;

INSERT INTO users (user_id, username, team_name, is_active) VALUES
    ('u1', 'Aleksandr', 'backend', true),
    ('u2', 'Dmitriy', 'backend', true),
    ('u3', 'Sergey', 'backend', true),
    ('u4', 'Mikhail', 'backend', false),
    
    ('u6', 'Ekaterina', 'frontend', true),
    ('u7', 'Olga', 'frontend', true),
    ('u8', 'Maria', 'frontend', true),
    
    ('u9', 'Vladimir', 'devops', true),
    ('u10', 'Nikolay', 'devops', true),
    ('u11', 'Ivan', 'devops', false),
    
    ('u12', 'Yulia', 'mobile', true),
    ('u13', 'Tatiana', 'mobile', true),
    ('u14', 'Pavel', 'mobile', true)
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at) VALUES
    ('pr-1001', 'Add authentication service', 'u1', 'OPEN', NOW() - INTERVAL '2 days'),
    ('pr-1002', 'Fix database migration', 'u2', 'MERGED', NOW() - INTERVAL '5 days'),
    ('pr-1003', 'Implement user dashboard', 'u5', 'OPEN', NOW() - INTERVAL '1 day'),
    ('pr-1004', 'Update deployment scripts', 'u9', 'OPEN', NOW() - INTERVAL '3 hours')
ON CONFLICT (pull_request_id) DO NOTHING;

INSERT INTO pr_reviewers (pull_request_id, user_id, assigned_at) VALUES
    ('pr-1001', 'u2', NOW() - INTERVAL '2 days'),
    ('pr-1001', 'u3', NOW() - INTERVAL '2 days'),
    ('pr-1002', 'u1', NOW() - INTERVAL '5 days'),
    ('pr-1002', 'u3', NOW() - INTERVAL '5 days'),
    ('pr-1003', 'u6', NOW() - INTERVAL '1 day'),
    ('pr-1003', 'u7', NOW() - INTERVAL '1 day'),
    ('pr-1004', 'u10', NOW() - INTERVAL '3 hours')
ON CONFLICT (pull_request_id, user_id) DO NOTHING;

UPDATE pull_requests SET merged_at = NOW() - INTERVAL '4 days' WHERE pull_request_id = 'pr-1002';
