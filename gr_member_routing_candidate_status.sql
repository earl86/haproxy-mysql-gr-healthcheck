

use sys;

DELIMITER $$

CREATE FUNCTION gr_viable_candidate()
RETURNS VARCHAR(3)
DETERMINISTIC
BEGIN
  RETURN (
    SELECT IF( MEMBER_STATE='ONLINE' AND ((SELECT COUNT(*) FROM performance_schema.replication_group_members WHERE MEMBER_STATE != 'ONLINE') >= ((SELECT COUNT(*) FROM performance_schema.replication_group_members)/2) = 0), 'YES', 'NO' )
    FROM performance_schema.replication_group_members WHERE MEMBER_ID=@@SERVER_UUID
  );
END $$

CREATE FUNCTION gr_read_only()
RETURNS VARCHAR(3)
DETERMINISTIC
BEGIN
  RETURN (
    SELECT IF( (SELECT (SELECT GROUP_CONCAT(variable_value) FROM performance_schema.global_variables WHERE variable_name IN ('read_only','super_read_only')) != 'OFF,OFF'), 'YES', 'NO')
  );
END $$

CREATE FUNCTION gr_transactions_behind()
RETURNS INT
DETERMINISTIC
BEGIN
  RETURN ( SELECT COUNT_TRANSACTIONS_REMOTE_IN_APPLIER_QUEUE FROM performance_schema.replication_group_member_stats WHERE MEMBER_ID=@@SERVER_UUID );
END $$

CREATE FUNCTION gr_transactions_to_cert()
RETURNS INT
DETERMINISTIC
BEGIN
  RETURN ( SELECT COUNT_TRANSACTIONS_IN_QUEUE FROM performance_schema.replication_group_member_stats WHERE MEMBER_ID=@@SERVER_UUID );
END $$


CREATE FUNCTION gr_member_role()
RETURNS  VARCHAR(32)
DETERMINISTIC
BEGIN
  RETURN ( SELECT MEMBER_ROLE FROM performance_schema.replication_group_members WHERE MEMBER_ID=@@SERVER_UUID );
END $$

CREATE FUNCTION gr_member_state()
RETURNS  VARCHAR(64)
DETERMINISTIC
BEGIN
  RETURN ( SELECT MEMBER_STATE FROM performance_schema.replication_group_members WHERE MEMBER_ID=@@SERVER_UUID );
END $$

DELIMITER ;

CREATE VIEW sys.gr_member_routing_candidate_status AS SELECT
sys.gr_viable_candidate() as viable_candidate,
sys.gr_read_only() as read_only,
sys.gr_transactions_behind() as transactions_behind,
sys.gr_transactions_to_cert() as transactions_to_cert,
sys.gr_member_role() as member_role,
sys.gr_member_state() as member_state;

select * from sys.gr_member_routing_candidate_status;
