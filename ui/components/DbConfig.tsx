"use client";

import React, { useState } from "react";
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
import { useUpdateCoreConfigMutation } from "@/lib/store/apis/configApi";
import { toast } from "sonner";

const DbConfigCard = () => {
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

  const handlePgChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPgConfig({ ...pgConfig, [e.target.name]: e.target.value });
  };

  const [updateCoreConfig] = useUpdateCoreConfigMutation();

  const handleSave = async () => {
    try {
      await updateCoreConfig({
        drop_excess_requests: true,
        prometheus_labels: ["env=prod"],
        allowed_origins: ["*"],
        initial_pool_size: 10,
        enable_logging: true,
        enable_governance: false,
        enforce_governance_header: false,
        allow_direct_keys: true,
      }).unwrap();
      toast.success("Core setting updated successfully.");
    } catch (err: any) {
      console.error("Failed to update config:", err);
      alert("Failed to update config");
    }
  };

  return (
    <Card className="rounded-2xl shadow-md">
      <CardHeader>
        <CardTitle>Database Configuration</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Select DB type */}
        <div>
          <Label className="mb-2">Database Type</Label>
          <Select value={dbType} onValueChange={(val) => setDbType(val as "sqlite" | "postgres")}>
            <SelectTrigger>
              <SelectValue placeholder="Choose database type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="sqlite">SQLite</SelectItem>
              <SelectItem value="postgres">Postgres</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {dbType === "sqlite" && (
          <div className="mt-4">
            <Label className="mb-2">SQLite Path</Label>
            <Input value={sqlitePath} onChange={(e) => setSqlitePath(e.target.value)} />
          </div>
        )}

        {dbType === "postgres" && (
          <>
            <div className="mt-4">
              <Label className="mb-2">Host</Label>
              <Input name="host" value={pgConfig.host} onChange={handlePgChange}  placeholder="localhost"/>
            </div>
            <div className="mt-4">
              <Label className="mb-2">Port</Label>
              <Input name="port" value={pgConfig.port} onChange={handlePgChange}
              placeholder="5432" />
            </div>
            <div className="mt-4">
              <Label className="mb-2">User</Label>
              <Input name="user" value={pgConfig.user} onChange={handlePgChange} 
              placeholder="user"/>
            </div>
            <div className="mt-4">
              <Label className="mb-2">Password</Label>
              <Input
                type="password"
                name="password"
                value={pgConfig.password}
                onChange={handlePgChange}
                placeholder="mypassword"
              />
            </div>
            <div className="mt-4">
              <Label className="mb-2">Database Name</Label>
              <Input name="dbname" value={pgConfig.dbname} onChange={handlePgChange} 
              placeholder="mydb"/>
            </div>
            <div className="mt-4">
              <Label className="mb-2">SSL Mode</Label>
              <Select
                value={pgConfig.sslmode}
                onValueChange={(val) => setPgConfig({ ...pgConfig, sslmode: val })}
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

        <Button className="w-full mt-6" onClick={handleSave}>
          Save Database Config
        </Button>
      </CardContent>
    </Card>
  );
};

export default DbConfigCard;
