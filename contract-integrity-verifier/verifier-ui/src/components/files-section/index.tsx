"use client";

import { useRef } from "react";
import { useVerifierStore } from "@/stores/verifier";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ALLOWED_SCHEMA_EXTENSIONS, ALLOWED_ARTIFACT_EXTENSIONS } from "@/lib/constants";
import styles from "./files-section.module.scss";

export function FilesSection() {
  const { requiredFiles, uploadedFiles, fileUploadErrors, uploadFile } = useVerifierStore();

  const inputRefs = useRef<Map<string, HTMLInputElement>>(new Map());

  if (requiredFiles.length === 0) {
    return null;
  }

  const handleFileSelect = (originalPath: string, file: File | null) => {
    if (file) {
      uploadFile(originalPath, file);
    }
  };

  const getAcceptedExtensions = (type: "schema" | "artifact") => {
    return type === "schema" ? ALLOWED_SCHEMA_EXTENSIONS : ALLOWED_ARTIFACT_EXTENSIONS;
  };

  const uploadedCount = requiredFiles.filter((f) => f.uploaded).length;
  const totalCount = requiredFiles.length;

  return (
    <Card
      title="Required Files"
      description={`Upload schema and artifact files referenced in your config (${uploadedCount}/${totalCount} uploaded)`}
    >
      <div className={styles.fileList}>
        {requiredFiles.map((file) => {
          const uploaded = uploadedFiles.get(file.path);
          const error = fileUploadErrors.get(file.path);
          const accepted = getAcceptedExtensions(file.type);

          return (
            <div key={file.path} className={styles.fileItem}>
              <div className={styles.fileInfo}>
                <span className={styles.fileType}>{file.type}</span>
                <div className={styles.fileDetails}>
                  <code className={styles.filePath}>{file.path}</code>
                  <span className={styles.fileContract}>Required by: {file.contractName}</span>
                </div>
              </div>

              <div className={styles.fileActions}>
                {uploaded ? (
                  <span className={styles.uploaded}>
                    <svg
                      className={styles.checkIcon}
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                      aria-hidden="true"
                    >
                      <polyline points="20 6 9 17 4 12" />
                    </svg>
                    {uploaded.filename}
                  </span>
                ) : (
                  <>
                    <input
                      type="file"
                      accept={accepted.join(",")}
                      onChange={(e) => handleFileSelect(file.path, e.target.files?.[0] || null)}
                      ref={(el) => {
                        if (el) inputRefs.current.set(file.path, el);
                      }}
                      className={styles.hiddenInput}
                      id={`file-${file.path}`}
                    />
                    <Button variant="secondary" size="sm" onClick={() => inputRefs.current.get(file.path)?.click()}>
                      Upload
                    </Button>
                  </>
                )}
              </div>

              {error && (
                <p className={styles.error} role="alert">
                  {error}
                </p>
              )}
            </div>
          );
        })}
      </div>
    </Card>
  );
}
