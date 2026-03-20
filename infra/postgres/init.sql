-- Initialize RentFlow databases

-- Create auth_service database
CREATE DATABASE auth_service OWNER postgres;

-- Create inventory_service database
CREATE DATABASE inventory_service OWNER postgres;

-- Create project_service database
CREATE DATABASE project_service OWNER postgres;

-- Create scanner_service database
CREATE DATABASE scanner_service OWNER postgres;

-- Create warehouse_service database
CREATE DATABASE warehouse_service OWNER postgres;

-- Create invoice_service database
CREATE DATABASE invoice_service OWNER postgres;

-- Create document_service database
CREATE DATABASE document_service OWNER postgres;

-- Create crew_service database
CREATE DATABASE crew_service OWNER postgres;

-- Create federation_service database
CREATE DATABASE federation_service OWNER postgres;

-- Create maintenance_service database
CREATE DATABASE maintenance_service OWNER postgres;

-- Create transport_service database
CREATE DATABASE transport_service OWNER postgres;

-- Create insurance_service database
CREATE DATABASE insurance_service OWNER postgres;

-- Create workflow_service database
CREATE DATABASE workflow_service OWNER postgres;

-- Create ai_service database
CREATE DATABASE ai_service OWNER postgres;

-- Create notification_service database
CREATE DATABASE notification_service OWNER postgres;

-- Create reporting_service database
CREATE DATABASE reporting_service OWNER postgres;

-- Create audit_service database
CREATE DATABASE audit_service OWNER postgres;

-- Create expense_service database
CREATE DATABASE expense_service OWNER postgres;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE auth_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE inventory_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE project_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE scanner_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE warehouse_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE invoice_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE document_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE crew_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE federation_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE maintenance_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE transport_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE insurance_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE workflow_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE ai_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE notification_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE reporting_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE audit_service TO postgres;
GRANT ALL PRIVILEGES ON DATABASE expense_service TO postgres;
