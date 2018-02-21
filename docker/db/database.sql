CREATE TABLE crash_reports (
      id INT NOT NULL AUTO_INCREMENT,
      plugin VARCHAR(128) DEFAULT NULL,
      version VARCHAR(127) DEFAULT NULL,
      build INT DEFAULT 0,
      file VARCHAR(255),
      message VARCHAR(255),
      line INT NOT NULL,
      type VARCHAR(32),
      os VARCHAR(16),
      reportType VARCHAR(128),
      submitDate INT,
      reportDate INT,
      duplicate BOOL,
      PRIMARY KEY (id),
      INDEX(plugin(10)),
      INDEX(message(32)),
      INDEX(file(32))
) ENGINE=MYISAM;
