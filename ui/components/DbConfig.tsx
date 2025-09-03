"use client";

import React, { useEffect, useState } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import {
  useGetCoreConfigQuery,
  useUpdateCoreConfigMutation,
} from "@/lib/store/apis/configApi";
import { BifrostConfig, CoreConfig, DatabaseConfig, PostgresConfig, SqliteConfig } from "@/lib/types/config";
import { toast } from "sonner";

const DbConfigCard = () => {
  const { data, isLoading, error } = useGetCoreConfigQuery({ fromDB: true });
  const [updateCoreConfig, { isLoading: isUpdating }] =
    useUpdateCoreConfigMutation();

  const [dbType, setDbType] = useState<"sqlite" | "postgres">("sqlite");
  const [sqlitePath, setSqlitePath] = useState("./bifrost.db");
  const [pgConfig, setPgConfig] = useState({
    host: "",
    port: "",
    user: "",
    password: "",
    dbname: "",
    sslmode: "disable",
  });
// ismei save changes wale button pr migration run krdo agr 
// changes h toh
  useEffect(() => {
    if (data) {
      const db = (data.client_config as any).db;

      if (db?.type === "sqlite") {
        setDbType("sqlite");
        setSqlitePath(db.sqlite_path || "./bifrost.db");
      } else if (db?.type === "postgres") {
        setDbType("postgres");
        setPgConfig({
          host: db.host || "",
          port: db.port || "",
          user: db.user || "",
          password: db.password || "",
          dbname: db.dbname || "",
          sslmode: db.sslmode || "disable",
        });
      }
    }
  }, [data]);

  const handlePgChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPgConfig({ ...pgConfig, [e.target.name]: e.target.value });
  };

  const handleSave = async () => {
    try {
      const sqliteConfig: SqliteConfig = {
        type: "sqlite",
        config: { path: sqlitePath },
      };

      const postgresConfig: PostgresConfig = {
        type: "postgres",
        config: {
          host: pgConfig.host,
          port: Number(pgConfig.port),
          user: pgConfig.user,
          password: pgConfig.password,
          database: pgConfig.dbname,
          sslMode:pgConfig.sslmode
        },
      };

      const dbConfig: DatabaseConfig = dbType === "sqlite" ? sqliteConfig : postgresConfig;

      const payload: CoreConfig = {
        ...data!.client_config,
        config_store: dbConfig,
      };

      await updateCoreConfig(payload).unwrap();

      toast.success("Database config updated successfully.");
    } catch (err: any) {
      console.error("Failed to update config:", err);
      toast.error("Failed to update config");
    }
  };

  if (isLoading) return <p>Loading config...</p>;
  if (error) return <p>Failed to load config</p>;

  return (
    <Card className="rounded-2xl shadow-md">
      <CardHeader>
        <CardTitle>Database Configuration</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* DB Type Selection */}
        <div>
          <Label className="mb-2">Database Type</Label>
          <Select
            value={dbType}
            onValueChange={(val) => setDbType(val as "sqlite" | "postgres")}
          >
            <SelectTrigger>
              <SelectValue placeholder="Choose database type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="sqlite">SQLite</SelectItem>
              <SelectItem value="postgres">Postgres</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* SQLite */}
        {dbType === "sqlite" && (
          <div className="mt-4">
            <Label className="mb-2">SQLite Path</Label>
            <Input
              value={sqlitePath}
              onChange={(e) => setSqlitePath(e.target.value)}
              placeholder="./bifrost.db"
            />
          </div>
        )}

        {/* Postgres */}
        {dbType === "postgres" && (
          <>
            {["host", "port", "user", "password", "dbname"].map((field) => (
              <div className="mt-4" key={field}>
                <Label className="mb-2">{field}</Label>
                <Input
                  type={field === "password" ? "password" : "text"}
                  name={field}
                  value={(pgConfig as any)[field]}
                  onChange={handlePgChange}
                  placeholder={field}
                />
              </div>
            ))}

            <div className="mt-4">
              <Label className="mb-2">SSL Mode</Label>
              <Select
                value={pgConfig.sslmode}
                onValueChange={(val) =>
                  setPgConfig({ ...pgConfig, sslmode: val })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select SSL Mode" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="disable">Disable</SelectItem>
                  <SelectItem value="enable">Enable</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </>
        )}

        <Button
          className="w-full mt-6"
          onClick={handleSave}
          disabled={isUpdating}
        >
          {isUpdating ? "Saving..." : "Save Database Config"}
        </Button>
      </CardContent>
    </Card>
  );
};

export default DbConfigCard;
