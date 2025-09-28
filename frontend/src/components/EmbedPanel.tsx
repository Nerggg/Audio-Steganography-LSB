"use client"

import type React from "react"
import { useState, useEffect } from "react"
import FileUpload from "./FileUpload"
import AudioPlayer from "./AudioPlayer"
import Button from "./Button"
import type { UploadedFile, EmbedOptions, EmbedResult, AppStatus, AudioCapacity } from "../types"

interface EmbedPanelProps {
  onStatusUpdate: (status: AppStatus) => void
  onCapacityUpdate: (capacity: AudioCapacity) => void
  onEmbedComplete: (result: EmbedResult) => void
}

const EmbedPanel: React.FC<EmbedPanelProps> = ({ onStatusUpdate, onCapacityUpdate, onEmbedComplete }) => {
  const [coverAudio, setCoverAudio] = useState<UploadedFile | undefined>()
  const [stegoAudio, setStegoAudio] = useState<UploadedFile | undefined>()
  const [secretFile, setSecretFile] = useState<UploadedFile | undefined>()
  const [options, setOptions] = useState<EmbedOptions>({
    stegKey: "",
    nLSB: 1,
    encrypt: false,
    randomStart: false,
  })
  const [isEmbedding, setIsEmbedding] = useState(false)

  // Calculate capacity when cover audio is loaded
  useEffect(() => {
    if (coverAudio) {
      // Simulate capacity calculation
      onStatusUpdate({
        isLoading: true,
        message: "Analyzing audio capacity...",
        type: "info",
      })

      setTimeout(() => {
        const capacity = {
          maxBytes: Math.floor(coverAudio.size * 0.1), // Simulate 10% capacity
          maxCharacters: Math.floor(coverAudio.size * 0.08), // Slightly less for characters
        }

        onCapacityUpdate(capacity)
        onStatusUpdate({
          isLoading: false,
          message: `Audio loaded. Capacity: ${capacity.maxCharacters} characters`,
          type: "success",
        })
      }, 1500)
    }
  }, [coverAudio, onStatusUpdate, onCapacityUpdate])

  const handleCoverAudioSelect = (file: UploadedFile) => {
    setCoverAudio(file)
    setStegoAudio(undefined) // Clear previous result
  }

  const handleSecretFileSelect = (file: UploadedFile) => {
    setSecretFile(file)
  }

  const handleEmbed = async () => {
    if (!coverAudio || !secretFile) return

    setIsEmbedding(true)
    onStatusUpdate({
      isLoading: true,
      message: "Embedding file into audio...",
      type: "info",
    })

    try {
      // Simulate embedding process
      await new Promise((resolve) => setTimeout(resolve, 3000))

      // Simulate creating stego audio (in real app, this would be the actual embedded audio)
      const stegoFile: UploadedFile = {
        file: coverAudio.file, // In real app, this would be the processed file
        name: `stego_${coverAudio.name}`,
        size: coverAudio.size,
        url: coverAudio.url, // In real app, this would be the new processed audio URL
      }

      setStegoAudio(stegoFile)

      // Simulate PSNR calculation
      const psnr = 35 + Math.random() * 10 // Random PSNR between 35-45 dB

      const result: EmbedResult = {
        success: true,
        psnr,
        stegoAudioUrl: stegoFile.url,
        message: "File successfully embedded",
      }

      onEmbedComplete(result)
    } catch (error) {
      const result: EmbedResult = {
        success: false,
        message: "Failed to embed file",
      }
      onEmbedComplete(result)
    } finally {
      setIsEmbedding(false)
    }
  }

  const isEmbedReady = coverAudio && secretFile && (options.encrypt || options.randomStart ? options.stegKey.trim() : true)

  const getKeyPlaceholder = () => {
    if (options.encrypt && options.randomStart) {
      return "Enter secret key for encryption and random seed..."
    } else if (options.encrypt) {
      return "Enter secret key for encryption..."
    } else if (options.randomStart) {
      return "Enter secret key for random seed..."
    }
    return ""
  }

  return (
    <div className="space-y-8">
      {/* Cover Audio Upload */}
      <div className="border border-purple-400/50 rounded-lg p-6 bg-purple-900/10 backdrop-blur-sm">
        <FileUpload
          onFileSelect={handleCoverAudioSelect}
          accept="audio/*"
          label="COVER AUDIO FILE"
          currentFile={coverAudio}
        />
      </div>

      {/* Secret File Upload */}
      <div className="border border-cyan-400/50 rounded-lg p-6 bg-black/30 backdrop-blur-sm">
        <FileUpload
          onFileSelect={handleSecretFileSelect}
          accept="*/*"
          label="SECRET FILE"
          currentFile={secretFile}
        />
      </div>

      {/* Advanced Options */}
      <div className="border border-yellow-400/50 rounded-lg p-6 bg-yellow-900/10 backdrop-blur-sm">
        <div className="text-yellow-400 font-mono text-sm mb-4 flex items-center gap-2">
          <span className="w-2 h-2 bg-yellow-400 rounded-full animate-pulse"></span>
          ADVANCED OPTIONS
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={options.encrypt}
              onChange={(e) => setOptions({ ...options, encrypt: e.target.checked })}
              className="w-5 h-5 text-yellow-400 bg-black border-yellow-400 rounded focus:ring-yellow-400 focus:ring-2"
            />
            <span className="text-white font-mono">VIGENERE CIPHER ENCRYPTION</span>
          </label>

          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={options.randomStart}
              onChange={(e) => setOptions({ ...options, randomStart: e.target.checked })}
              className="w-5 h-5 text-yellow-400 bg-black border-yellow-400 rounded focus:ring-yellow-400 focus:ring-2"
            />
            <span className="text-white font-mono">RANDOM START POSITION</span>
          </label>
        </div>

        {(options.encrypt || options.randomStart) && (
          <div className="mt-6 border border-pink-500/50 rounded-lg p-6 bg-pink-900/10 backdrop-blur-sm">
            <div className="text-pink-400 font-mono text-sm mb-4 flex items-center gap-2">
              <span className="w-2 h-2 bg-pink-400 rounded-full animate-pulse"></span>
              SECRET KEY
            </div>

            <input
              type="password"
              value={options.stegKey}
              onChange={(e) => setOptions({ ...options, stegKey: e.target.value })}
              placeholder={getKeyPlaceholder()}
              className="w-full bg-black/50 border border-pink-400/30 rounded-lg p-3 text-white font-mono focus:border-pink-400 focus:outline-none focus:ring-2 focus:ring-pink-400/25 transition-all"
            />
          </div>
        )}
      </div>

      {/* LSB Options */}
      <div className="border border-green-400/50 rounded-lg p-6 bg-green-900/10 backdrop-blur-sm">
        <div className="text-green-400 font-mono text-sm mb-4 flex items-center gap-2">
          <span className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
          LSB CONFIGURATION
        </div>

        <div className="space-y-3">
          <div>
            <label className="text-white font-mono text-sm mb-2 block">N-LSB BITS:</label>
            <div className="flex gap-4">
              {[1, 2, 4].map((bits) => (
                <label key={bits} className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="nLSB"
                    value={bits}
                    checked={options.nLSB === bits}
                    onChange={(e) => setOptions({ ...options, nLSB: Number.parseInt(e.target.value) as 1 | 2 | 4 })}
                    className="w-4 h-4 text-green-400 bg-black border-green-400 focus:ring-green-400 focus:ring-2"
                  />
                  <span className="text-white font-mono">{bits}</span>
                </label>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Embed Button */}
      <div className="flex justify-center">
        <Button
          onClick={handleEmbed}
          disabled={!isEmbedReady || isEmbedding}
          variant="primary"
          className="px-12 py-4 text-lg"
        >
          {isEmbedding ? "EMBEDDING..." : "EMBED FILE"}
        </Button>
      </div>

      {/* Audio Players */}
      {coverAudio && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <AudioPlayer audioUrl={coverAudio.url} label="ORIGINAL AUDIO" />
          {stegoAudio && <AudioPlayer audioUrl={stegoAudio.url} label="STEGO AUDIO" />}
        </div>
      )}
    </div>
  )
}

export default EmbedPanel
