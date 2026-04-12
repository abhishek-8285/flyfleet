import { PaperClipIcon, XMarkIcon } from "@heroicons/react/24/outline";
import { useCallback, useState } from "react";

interface TripDocument {
  file: File;
  preview: string;
  type: "image" | "pdf" | "other";
}

interface DocumentUploaderProps {
  maxFiles?: number;
  maxSizeMB?: number;
  onChange: (files: TripDocument[]) => void;
}

export default function DocumentUploader({
  maxFiles = 5,
  maxSizeMB = 10,
  onChange,
}: DocumentUploaderProps) {
  const [documents, setDocuments] = useState<TripDocument[]>([]);
  const [dragOver, setDragOver] = useState(false);
  const [error, setError] = useState("");

  const handleFiles = useCallback(
    (fileList: FileList | File[]) => {
      setError("");
      const remaining = maxFiles - documents.length;
      if (remaining <= 0) {
        setError(`Maximum ${maxFiles} files allowed`);
        return;
      }

      const maxSize = maxSizeMB * 1024 * 1024;
      const newDocs: TripDocument[] = [];

      Array.from(fileList)
        .slice(0, remaining)
        .forEach((file) => {
          if (file.size > maxSize) {
            setError(`${file.name} exceeds ${maxSizeMB}MB limit`);
            return;
          }

          const isImage = file.type.startsWith("image/");
          const isPdf = file.type === "application/pdf";
          const type = isImage ? "image" : isPdf ? "pdf" : "other";

          let preview = "";
          if (isImage) {
            preview = URL.createObjectURL(file);
          }

          newDocs.push({ file, preview, type });
        });

      const updated = [...documents, ...newDocs];
      setDocuments(updated);
      onChange(updated);
    },
    [documents, maxFiles, maxSizeMB, onChange],
  );

  const removeDocument = (index: number) => {
    const updated = [...documents];
    if (updated[index].preview) {
      URL.revokeObjectURL(updated[index].preview);
    }
    updated.splice(index, 1);
    setDocuments(updated);
    onChange(updated);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    handleFiles(e.dataTransfer.files);
  };

  return (
    <div className="space-y-3">
      {/* Drop Zone */}
      <div
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        className={`border-2 border-dashed rounded-xl p-6 text-center transition-colors cursor-pointer ${
          dragOver
            ? "border-primary bg-primary/5"
            : "border-border-color hover:border-primary/40"
        }`}
        onClick={() =>
          document.getElementById("doc-upload-input")?.click()
        }
      >
        <input
          id="doc-upload-input"
          type="file"
          multiple
          accept="image/*,.pdf"
          className="hidden"
          onChange={(e) => e.target.files && handleFiles(e.target.files)}
        />
        <PaperClipIcon className="w-8 h-8 text-text-muted mx-auto mb-2" />
        <p className="text-sm text-white font-medium">
          Drop files here or{" "}
          <span className="text-primary underline">browse</span>
        </p>
        <p className="text-xs text-text-muted mt-1">
          Images (PNG, JPG, WEBP) or PDF • Max {maxSizeMB}MB each • Up to {maxFiles} files
        </p>
        {error && (
          <p className="text-xs text-danger mt-2 flex items-center justify-center gap-1">
            <span className="material-symbols-outlined text-sm">error</span>
            {error}
          </p>
        )}
      </div>

      {/* File List */}
      {documents.length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
          {documents.map((doc, i) => (
            <div
              key={i}
              className="relative group bg-surface-darker border border-border-color rounded-lg overflow-hidden"
            >
              {/* Preview */}
              <div className="aspect-square bg-surface-dark flex items-center justify-center overflow-hidden">
                {doc.type === "image" && doc.preview ? (
                  <img
                    src={doc.preview}
                    alt={doc.file.name}
                    className="w-full h-full object-cover"
                  />
                ) : doc.type === "pdf" ? (
                  <div className="flex flex-col items-center gap-1">
                    <span className="material-symbols-outlined text-3xl text-red-400">
                      picture_as_pdf
                    </span>
                    <span className="text-[10px] text-text-muted">PDF</span>
                  </div>
                ) : (
                  <div className="flex flex-col items-center gap-1">
                    <span className="material-symbols-outlined text-3xl text-text-muted">
                      description
                    </span>
                    <span className="text-[10px] text-text-muted">File</span>
                  </div>
                )}
              </div>

              {/* File info */}
              <div className="px-2 py-1.5 border-t border-border-color">
                <p className="text-[10px] text-white truncate" title={doc.file.name}>
                  {doc.file.name}
                </p>
                <p className="text-[9px] text-text-muted">
                  {(doc.file.size / 1024).toFixed(0)} KB
                </p>
              </div>

              {/* Remove button */}
              <button
                type="button"
                onClick={() => removeDocument(i)}
                className="absolute top-1 right-1 bg-surface-darker/90 text-text-muted hover:text-danger p-0.5 rounded opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <XMarkIcon className="w-3.5 h-3.5" />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
