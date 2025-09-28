"use client"

import type React from "react"
import { useState } from "react"
import FileUpload from "./FileUpload"
import AudioPlayer from "./AudioPlayer"
import Button from "./Button"
import type { UploadedFile, ExtractOptions, ExtractResult, AppStatus } from "../types"

interface ExtractPanelProps {
  onStatusUpdate: (status: AppStatus) => void
  onExtractComplete: (result: ExtractResult) => void
}

const ExtractPanel: React.FC<ExtractPanelProps> = ({ onStatusUpdate, onExtractComplete }) => {
  const [stegoAudio, setStegoAudio] = useState<UploadedFile | undefined>()
  const [options, setOptions] = useState<ExtractOptions>({
    stegKey: "",
  })
  const [isExtracting, setIsExtracting] = useState(false)
  const [extractedFileInfo, setExtractedFileInfo] = useState<{
    filename: string
    size: number
    blobUrl: string
  } | null>(null)

  const API_URL = "http://localhost:8080"

  const handleStegoAudioSelect = (file: UploadedFile) => {
    setStegoAudio(file)
    setExtractedFileInfo(null)
  }

  const handleExtract = async () => {
    if (!stegoAudio) return

    setIsExtracting(true)
    setExtractedFileInfo(null)
    onStatusUpdate({
      isLoading: true,
      message: "Analyzing steganographic audio...",
      type: "info",
    })

    try {
      const formData = new FormData()
      formData.append("stego_audio", stegoAudio.file)
      if (options.stegKey) {
        formData.append("stego_key", options.stegKey)
      }

      const response = await fetch(`${API_URL}/api/v1/extract`, {
        method: "POST",
        body: formData,
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || "Extraction failed")
      }

      const contentDisposition = response.headers.get("Content-Disposition")
      const filenameMatch = contentDisposition?.match(/filename="(.+)"/)
      const filename = filenameMatch ? filenameMatch[1] : "extracted_file"
      const secretSize = parseInt(response.headers.get("X-Secret-Size") || "0")

      const blob = await response.blob()
      const blobUrl = URL.createObjectURL(blob)

      setExtractedFileInfo({
        filename,
        size: secretSize,
        blobUrl,
      })

      const result: ExtractResult = {
        success: true,
        message: "File extracted successfully",
      }
      onExtractComplete(result)
    } catch (error) {
      const result: ExtractResult = {
        success: false,
        error: error instanceof Error ? error.message : "Failed to extract file",
      }
      onExtractComplete(result)
    } finally {
      setIsExtracting(false)
    }
  }

  const isExtractReady = stegoAudio

  return (
    <div className="space-y-8">
      {/* Stego Audio Upload */}
      <div className="border border-purple-400/50 rounded-lg p-6 bg-purple-900/10 backdrop-blur-sm">
        <FileUpload
          onFileSelect={handleStegoAudioSelect}
          accept="audio/*"
          label="STEGANOGRAPHIC AUDIO FILE"
          currentFile={stegoAudio}
        />
      </div>

      {/* Optional Stego Key */}
      <div className="border border-pink-500/50 rounded-lg p-6 bg-pink-900/10 backdrop-blur-sm">
        <div className="text-pink-400 font-mono text-sm mb-4 flex items-center gap-2">
          <span className="w-2 h-2 bg-pink-400 rounded-full animate-pulse"></span>
          SECRET KEY (OPTIONAL)
        </div>

        <input
          type="password"
          value={options.stegKey}
          onChange={(e) => setOptions({ ...options, stegKey: e.target.value })}
          placeholder="Enter decryption key if message is encrypted or using random start position..."
          className="w-full bg-black/50 border border-pink-400/30 rounded-lg p-3 text-white font-mono focus:border-pink-400 focus:outline-none focus:ring-2 focus:ring-pink-400/25 transition-all"
        />
      </div>

      {/* Extract Button */}
      <div className="flex justify-center">
        <Button
          onClick={handleExtract}
          disabled={!isExtractReady || isExtracting}
          variant="primary"
          className="px-12 py-4 text-lg"
        >
          {isExtracting ? "EXTRACTING..." : "EXTRACT MESSAGE"}
        </Button>
      </div>

      {/* Extracted File Info Display */}
      {extractedFileInfo && (
        <div className="border border-green-400/50 rounded-lg p-6 bg-green-900/10 backdrop-blur-sm">
          <div className="text-green-400 font-mono text-sm mb-4 flex items-center gap-2">
            <span className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
            EXTRACTED FILE
          </div>

          <div className="bg-black/70 border border-green-400/30 rounded-lg p-4 min-h-[120px] relative overflow-hidden">
            <div className="flex items-center gap-2 mb-3 pb-2 border-b border-green-400/20">
              <div className="w-3 h-3 rounded-full bg-red-500"></div>
              <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
              <div className="w-3 h-3 rounded-full bg-green-500"></div>
              <span className="text-green-400 font-mono text-xs ml-2">{extractedFileInfo.filename}</span>
            </div>

            <div className="font-mono text-green-400 leading-relaxed">
              <p>File Name: {extractedFileInfo.filename}</p>
              <p>Size: {(extractedFileInfo.size / 1024).toFixed(2)} KB</p>
            </div>

            <div className="absolute inset-0 pointer-events-none">
              <div className="absolute w-full h-px bg-gradient-to-r from-transparent via-green-400/30 to-transparent animate-pulse"></div>
            </div>
          </div>

          <div className="flex justify-between items-center mt-4">
            <div className="flex items-center gap-2 text-sm">
              <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
              <span className="text-green-400 font-mono">FILE EXTRACTED</span>
            </div>

            <div className="flex gap-2">
              <a
                href={extractedFileInfo.blobUrl}
                download={extractedFileInfo.filename}
                className="px-3 py-1 border border-green-400/50 rounded text-green-400 text-sm font-mono hover:bg-green-400/10 transition-colors"
              >
                DOWNLOAD
              </a>
            </div>
          </div>
        </div>
      )}

      {/* Audio Player */}
      {stegoAudio && (
        <div className="max-w-2xl mx-auto">
          <AudioPlayer audioUrl={stegoAudio.url} label="STEGANOGRAPHIC AUDIO" />
        </div>
      )}
    </div>
  )
}

export default ExtractPanel
