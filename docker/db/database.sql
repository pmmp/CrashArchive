SET GLOBAL INNODB_FILE_PER_TABLE=1;
SET GLOBAL INNODB_FILE_FORMAT=Barracuda;

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
      submitDate INT,
      reportDate INT,
      duplicate BOOL,
      reporterName VARCHAR(127),
      reporterEmail VARCHAR(127),
      PRIMARY KEY (id),
      INDEX(plugin(10)),
      INDEX(message(32)),
      INDEX(file(32))
) ENGINE=INNODB;

CREATE TABLE crash_report_blobs (
      id INT NOT NULL,
      crash_report_json MEDIUMBLOB NOT NULL,
      PRIMARY KEY (id),
      FOREIGN KEY (id)
            REFERENCES crash_reports(id)
            ON DELETE CASCADE
) ENGINE=INNODB ROW_FORMAT=COMPRESSED;