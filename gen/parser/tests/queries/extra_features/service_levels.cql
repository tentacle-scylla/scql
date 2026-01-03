-- ScyllaDB Service Level Management (QoS)
-- https://docs.scylladb.com/stable/cql/service-levels.html

-- Create service levels
CREATE SERVICE LEVEL gold;
CREATE SERVICE LEVEL silver WITH timeout = 500ms;
CREATE SERVICE LEVEL bronze WITH timeout = 1s AND workload_type = 'batch';
CREATE SERVICE LEVEL IF NOT EXISTS premium;
CREATE SERVICE LEVEL enterprise WITH shares = 1000;

-- Alter service levels
ALTER SERVICE LEVEL gold WITH timeout = 100ms;
ALTER SERVICE LEVEL silver WITH shares = 500;
ALTER SERVICE LEVEL bronze WITH workload_type = 'interactive';

-- Drop service levels
DROP SERVICE LEVEL gold;
DROP SERVICE LEVEL IF EXISTS silver;

-- Attach/Detach service levels to roles
ATTACH SERVICE LEVEL gold TO admin;
ATTACH SERVICE LEVEL silver TO readonly_user;
DETACH SERVICE LEVEL FROM admin;

-- List service levels
LIST SERVICE LEVEL gold;
LIST ALL SERVICE LEVELS;
LIST ATTACHED SERVICE LEVEL OF admin;
LIST EFFECTIVE SERVICE LEVEL OF admin;
