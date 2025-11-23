\c prservice;

INSERT INTO teams (name) VALUES
('Backend Team'),
('Frontend Team'),
('Data Science'),
('DevOps'),
('Mobile Team'),
('Analytics'),
('QA Team'),
('Integration'),
('Security'),
('Infrastructure');

INSERT INTO users (user_id, name, is_active, team_id) VALUES
('u001', 'Alice', true, 1),
('u002', 'Bob', true, 1),
('u003', 'Charlie', true, 1),
('u004', 'David', true, 1),
('u005', 'Eva', false, 1),

('u006', 'Frank', true, 2),
('u007', 'Grace', true, 2),
('u008', 'Helen', true, 2),
('u009', 'Ian', false, 2),
('u010', 'Jane', true, 2),

('u011', 'Kyle', true, 3),
('u012', 'Laura', true, 3),
('u013', 'Mike', true, 3),
('u014', 'Nina', true, 3),
('u015', 'Oscar', false, 3),

('u016', 'Paul', true, 4),
('u017', 'Quinn', true, 4),
('u018', 'Rachel', true, 4),
('u019', 'Steve', false, 4),
('u020', 'Tina', true, 4),

('u021', 'Uma', true, 5),
('u022', 'Victor', true, 5),
('u023', 'Wendy', true, 5),
('u024', 'Xavier', false, 5),
('u025', 'Yara', true, 5),

('u026', 'Zack', true, 6),
('u027', 'Anna', true, 6),
('u028', 'Ben', true, 6),
('u029', 'Cody', true, 6),
('u030', 'Dina', false, 6),

('u031', 'Evan', true, 7),
('u032', 'Fiona', true, 7),
('u033', 'Gina', true, 7),
('u034', 'Hank', true, 7),
('u035', 'Iris', false, 7),

('u036', 'Jack', true, 8),
('u037', 'Kelly', true, 8),
('u038', 'Leo', true, 8),
('u039', 'Mona', false, 8),
('u040', 'Nick', true, 8),

('u041', 'Olga', true, 9),
('u042', 'Peter', true, 9),
('u043', 'Queen', true, 9),
('u044', 'Roger', true, 9),
('u045', 'Sara', false, 9),

('u046', 'Tim', true, 10),
('u047', 'Ursula', true, 10),
('u048', 'Vlad', true, 10),
('u049', 'Will', false, 10),
('u050', 'Zoe', true, 10);

INSERT INTO pull_requests (pr_id, name, author_id, pr_status, merged_at) VALUES
('pr001', 'Refactor backend module', 'u001', 'MERGED', '2024-11-10 12:00:00'),
('pr002', 'Fix auth bug', 'u003', 'OPEN', NULL),
('pr003', 'Improve logging', 'u004', 'MERGED', '2024-11-12 09:30:00'),
('pr004', 'Add caching', 'u002', 'OPEN', NULL),

('pr005', 'UI redesign', 'u006', 'MERGED', '2024-10-04 15:10:00'),
('pr006', 'Fix button styles', 'u008', 'OPEN', NULL),
('pr007', 'Add dark mode', 'u010', 'MERGED', '2024-11-03 10:00:00'),

('pr008', 'ML training pipeline', 'u011', 'OPEN', NULL),
('pr009', 'Optimize inference', 'u013', 'MERGED', '2024-09-19 14:00:00'),

('pr010', 'Kubernetes config update', 'u016', 'MERGED', '2024-08-17 17:40:00'),
('pr011', 'CI/CD improvements', 'u018', 'OPEN', NULL),

('pr012', 'iOS build fix', 'u021', 'MERGED', '2024-11-01 11:55:00'),
('pr013', 'Android analytics event fix', 'u023', 'OPEN', NULL),

('pr014', 'ETL improvement', 'u026', 'MERGED', '2024-10-10 13:20:00'),
('pr015', 'New dashboard page', 'u028', 'MERGED', '2024-11-05 18:25:00'),

('pr016', 'Test automation update', 'u031', 'OPEN', NULL),
('pr017', 'Improve QA coverage', 'u033', 'MERGED', '2024-09-28 07:30:00'),

('pr018', 'API integration update', 'u036', 'OPEN', NULL),
('pr019', 'Service connector fixes', 'u038', 'MERGED', '2024-07-21 09:10:00'),

('pr020', 'Firewall rules update', 'u041', 'MERGED', '2024-11-06 12:00:00');

INSERT INTO pull_requests_reviewers (pr_id, reviewer_id) VALUES
('pr001', 'u002'),
('pr001', 'u003'),

('pr002', 'u001'),

('pr003', 'u005'),

('pr004', 'u003'),
('pr004', 'u005'),

('pr005', 'u007'),
('pr006', 'u009'),

('pr007', 'u006'),
('pr007', 'u009'),

('pr008', 'u012'),

('pr009', 'u014'),
('pr009', 'u015'),

('pr010', 'u017'),
('pr011', 'u019'),

('pr012', 'u022'),
('pr013', 'u024'),

('pr014', 'u027'),
('pr015', 'u029'),

('pr016', 'u032'),
('pr017', 'u034'),

('pr018', 'u037'),
('pr019', 'u039'),

('pr020', 'u042'),
('pr020', 'u044');
