"use client"

import type React from "react"
import { useRef, useState } from "react"
import type { FileUploadProps } from "../types"

const FileUpload: React.FC<FileUploadProps> = ({ onFileSelect, accept = "audio/*", label, currentFile }) => {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [isDragOver, setIsDragOver] = useState(false)
  const [isProcessing, setIsProcessing] = useState(false)

  const handleFileSelect = async (file: File) => {
    if (!file) return

    setIsProcessing(true)

    try {
      // Create object URL for the file
      const url = URL.createObjectURL(file)

      // Create UploadedFile object
      const uploadedFile = {
        file,
        name: file.name,
        size: file.size,
        url,
      }

      // Simulate processing delay for better UX
      await new Promise((resolve) => setTimeout(resolve, 500))

      onFileSelect(uploadedFile)
    } catch (error) {
      console.error("Error processing file:", error)
    } finally {
      setIsProcessing(false)
    }
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      handleFileSelect(file)
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(false)

    const file = e.dataTransfer.files[0]
    if ((label == "COVER AUDIO FILE" || label == "STEGANOGRAPHIC AUDIO FILE") && file && file.type.startsWith("audio/")) {
      handleFileSelect(file)
    } else if (label == "SECRET FILE") {
      handleFileSelect(file)
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(true)
  }

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(false)
  }

  const openFileDialog = () => {
    fileInputRef.current?.click()
  }

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes"
    const k = 1024
    const sizes = ["Bytes", "KB", "MB", "GB"]
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${Number.parseFloat((bytes / k ** i).toFixed(2))} ${sizes[i]}`
  }

  return (
    <div className="space-y-4">
      {/* Label */}
      <div className="text-cyan-400 font-mono text-sm flex items-center gap-2">
        <span className="w-2 h-2 bg-cyan-400 rounded-full animate-pulse"></span>
        {label}
      </div>

      {/* Upload area */}
      <div
        className={`
          relative border-2 border-dashed rounded-lg p-6 transition-all duration-300 cursor-pointer
          ${
            isDragOver
              ? "border-pink-500 bg-pink-500/10 shadow-lg shadow-pink-500/25"
              : currentFile
                ? "border-green-400 bg-green-400/10"
                : "border-cyan-400/50 bg-black/30 hover:border-cyan-400 hover:bg-cyan-400/5"
          }
          ${isProcessing ? "animate-pulse" : ""}
        `}
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onClick={openFileDialog}
      >
        <input ref={fileInputRef} type="file" accept={accept} onChange={handleInputChange} className="hidden" />

        {isProcessing ? (
          <div className="text-center">
            <div className="flex justify-center items-center gap-2 mb-2">
              <div className="w-3 h-3 bg-cyan-400 rounded-full animate-ping"></div>
              <div className="w-3 h-3 bg-pink-500 rounded-full animate-ping animation-delay-200"></div>
              <div className="w-3 h-3 bg-purple-500 rounded-full animate-ping animation-delay-400"></div>
            </div>
            <p className="text-cyan-400 font-mono">PROCESSING FILE...</p>
          </div>
        ) : currentFile ? (
          <div className="text-center">
            <p className="text-green-400 font-mono font-bold mb-1">{currentFile.name}</p>
            <p className="text-gray-400 text-sm">{formatFileSize(currentFile.size)}</p>
            <div className="mt-3 flex items-center justify-center gap-2">
              <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
              <span className="text-green-400 text-xs font-mono">FILE LOADED</span>
            </div>
          </div>
        ) : (
          <div className="text-center">
            <div className="text-4xl mb-4 opacity-50">📁</div>
            <p className="text-cyan-400 font-mono mb-2">DROP {label} HERE</p>
            <p className="text-gray-400 text-sm mb-4">or click to browse</p>
            <div className="inline-block px-4 py-2 border border-cyan-400/50 rounded text-cyan-400 text-sm font-mono hover:bg-cyan-400/10 transition-colors">
              SELECT FILE
            </div>
          </div>
        )}

        {/* Corner decorations */}
        <div className="absolute top-2 left-2 w-4 h-4 border-l-2 border-t-2 border-current opacity-30"></div>
        <div className="absolute top-2 right-2 w-4 h-4 border-r-2 border-t-2 border-current opacity-30"></div>
        <div className="absolute bottom-2 left-2 w-4 h-4 border-l-2 border-b-2 border-current opacity-30"></div>
        <div className="absolute bottom-2 right-2 w-4 h-4 border-r-2 border-b-2 border-current opacity-30"></div>

        {/* Scanning line effect when processing */}
        {isProcessing && (
          <div className="absolute inset-0 overflow-hidden pointer-events-none">
            <div className="absolute w-full h-0.5 bg-gradient-to-r from-transparent via-cyan-400 to-transparent animate-pulse"></div>
          </div>
        )}
      </div>

      {/* File info panel */}
      {currentFile && (
        <div className="border border-cyan-400/30 rounded-lg p-3 bg-black/20 backdrop-blur-sm">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-cyan-400 font-mono">NAME:</span>
              <div className="text-white truncate">{currentFile.name}</div>
            </div>
            <div>
              <span className="text-cyan-400 font-mono">SIZE:</span>
              <div className="text-white">{formatFileSize(currentFile.size)}</div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default FileUpload
