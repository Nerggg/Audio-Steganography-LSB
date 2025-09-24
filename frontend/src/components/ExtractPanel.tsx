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
  const [extractedMessage, setExtractedMessage] = useState("")
  const [options, setOptions] = useState<ExtractOptions>({
    stegKey: "",
  })
  const [isExtracting, setIsExtracting] = useState(false)
  const [showMessage, setShowMessage] = useState(false)

  const handleStegoAudioSelect = (file: UploadedFile) => {
    setStegoAudio(file)
    setExtractedMessage("")
    setShowMessage(false)
  }

  const handleExtract = async () => {
    if (!stegoAudio) return

    setIsExtracting(true)
    setShowMessage(false)
    onStatusUpdate({
      isLoading: true,
      message: "Analyzing steganographic audio...",
      type: "info",
    })

    try {
      // Simulate extraction process
      await new Promise((resolve) => setTimeout(resolve, 2000))

      // Simulate message extraction (in real app, this would be actual extraction)
      const messages = [
        "The secret meeting is at midnight in the old warehouse.",
        "Operation Blue Moon is a go. Proceed with phase 2.",
        "The password is: CyberSteg2024!",
        "Hidden data successfully extracted from audio stream.",
        "Confidential: Project X files are in the secure vault.",
      ]

      const randomMessage = messages[Math.floor(Math.random() * messages.length)]
      setExtractedMessage(randomMessage)

      // Simulate typing effect
      setTimeout(() => {
        setShowMessage(true)
      }, 500)

      const result: ExtractResult = {
        success: true,
        message: randomMessage,
      }

      onExtractComplete(result)
    } catch (error) {
      const result: ExtractResult = {
        success: false,
        error: "Failed to extract message from audio",
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
          STEGANOGRAPHY KEY (OPTIONAL)
        </div>

        <input
          type="password"
          value={options.stegKey}
          onChange={(e) => setOptions({ ...options, stegKey: e.target.value })}
          placeholder="Enter decryption key if message is encrypted..."
          className="w-full bg-black/50 border border-pink-400/30 rounded-lg p-3 text-white font-mono focus:border-pink-400 focus:outline-none focus:ring-2 focus:ring-pink-400/25 transition-all"
        />

        <p className="text-gray-400 text-sm mt-2 font-mono">
          Leave empty if the message was not encrypted or if you want to attempt extraction without a key.
        </p>
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

      {/* Extracted Message Display */}
      {extractedMessage && (
        <div className="border border-green-400/50 rounded-lg p-6 bg-green-900/10 backdrop-blur-sm">
          <div className="text-green-400 font-mono text-sm mb-4 flex items-center gap-2">
            <span className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
            EXTRACTED MESSAGE
          </div>

          <div className="bg-black/70 border border-green-400/30 rounded-lg p-4 min-h-[120px] relative overflow-hidden">
            {/* Terminal-style header */}
            <div className="flex items-center gap-2 mb-3 pb-2 border-b border-green-400/20">
              <div className="w-3 h-3 rounded-full bg-red-500"></div>
              <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
              <div className="w-3 h-3 rounded-full bg-green-500"></div>
              <span className="text-green-400 font-mono text-xs ml-2">DECRYPTED_MESSAGE.TXT</span>
            </div>

            {/* Message content with typing effect */}
            <div className="font-mono text-green-400 leading-relaxed">
              {showMessage ? (
                <div className="terminal-text">{extractedMessage}</div>
              ) : (
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-green-400 rounded-full animate-ping"></div>
                  <span>Decrypting message...</span>
                </div>
              )}
            </div>

            {/* Cursor blink */}
            {showMessage && <span className="inline-block w-2 h-5 bg-green-400 ml-1 animate-pulse"></span>}

            {/* Scanning lines effect */}
            <div className="absolute inset-0 pointer-events-none">
              <div className="absolute w-full h-px bg-gradient-to-r from-transparent via-green-400/30 to-transparent animate-pulse"></div>
            </div>
          </div>

          {/* Message actions */}
          <div className="flex justify-between items-center mt-4">
            <div className="flex items-center gap-2 text-sm">
              <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
              <span className="text-green-400 font-mono">MESSAGE DECODED</span>
            </div>

            <div className="flex gap-2">
              <button
                onClick={() => navigator.clipboard.writeText(extractedMessage)}
                className="px-3 py-1 border border-green-400/50 rounded text-green-400 text-sm font-mono hover:bg-green-400/10 transition-colors"
              >
                COPY
              </button>
              <button
                onClick={() => {
                  const blob = new Blob([extractedMessage], { type: "text/plain" })
                  const url = URL.createObjectURL(blob)
                  const a = document.createElement("a")
                  a.href = url
                  a.download = "extracted_message.txt"
                  a.click()
                  URL.revokeObjectURL(url)
                }}
                className="px-3 py-1 border border-green-400/50 rounded text-green-400 text-sm font-mono hover:bg-green-400/10 transition-colors"
              >
                SAVE
              </button>
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

      {/* Extraction Info Panel */}
      <div className="border border-cyan-400/50 rounded-lg p-6 bg-cyan-900/10 backdrop-blur-sm">
        <div className="text-cyan-400 font-mono text-sm mb-4 flex items-center gap-2">
          <span className="w-2 h-2 bg-cyan-400 rounded-full animate-pulse"></span>
          EXTRACTION PARAMETERS
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
          <div>
            <span className="text-cyan-400 font-mono">METHOD:</span>
            <div className="text-white">LSB Analysis</div>
          </div>
          <div>
            <span className="text-cyan-400 font-mono">ENCRYPTION:</span>
            <div className="text-white">Auto-Detect</div>
          </div>
          <div>
            <span className="text-cyan-400 font-mono">STATUS:</span>
            <div className="text-white">{isExtracting ? "Processing..." : extractedMessage ? "Complete" : "Ready"}</div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ExtractPanel
