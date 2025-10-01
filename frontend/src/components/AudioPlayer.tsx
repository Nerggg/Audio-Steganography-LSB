"use client"

import type React from "react"
import { useRef, useState, useEffect } from "react"
import type { AudioPlayerProps } from "../types"
import Button from "./Button"
import { ArrowDown } from "lucide-react";

const AudioPlayer: React.FC<AudioPlayerProps> = ({ audioUrl, label, className = "", onClick }) => {
  const audioRef = useRef<HTMLAudioElement>(null)
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [volume, setVolume] = useState(1)

  useEffect(() => {
    // Add custom styles for range sliders
    const style = document.createElement('style')
    style.textContent = `
      .audio-player input[type="range"]::-webkit-slider-thumb {
        -webkit-appearance: none;
        appearance: none;
        height: 12px;
        width: 12px;
        border-radius: 50%;
        background: #ffffff;
        cursor: pointer;
        border: 2px solid #22d3ee;
        box-shadow: 0 0 4px rgba(34, 211, 238, 0.5);
      }
      
      .audio-player input[type="range"]::-moz-range-thumb {
        height: 12px;
        width: 12px;
        border-radius: 50%;
        background: #ffffff;
        cursor: pointer;
        border: 2px solid #22d3ee;
        box-shadow: 0 0 4px rgba(34, 211, 238, 0.5);
        -moz-appearance: none;
        appearance: none;
      }
    `
    document.head.appendChild(style)

    return () => {
      document.head.removeChild(style)
    }
  }, [])

  useEffect(() => {
    const audio = audioRef.current
    if (!audio) return

    const updateTime = () => setCurrentTime(audio.currentTime)
    const updateDuration = () => setDuration(audio.duration)
    const handleEnded = () => setIsPlaying(false)

    audio.addEventListener("timeupdate", updateTime)
    audio.addEventListener("loadedmetadata", updateDuration)
    audio.addEventListener("ended", handleEnded)

    return () => {
      audio.removeEventListener("timeupdate", updateTime)
      audio.removeEventListener("loadedmetadata", updateDuration)
      audio.removeEventListener("ended", handleEnded)
    }
  }, [audioUrl])

  const togglePlay = () => {
    const audio = audioRef.current
    if (!audio) return

    if (isPlaying) {
      audio.pause()
    } else {
      audio.play()
    }
    setIsPlaying(!isPlaying)
  }

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const audio = audioRef.current
    if (!audio) return

    const newTime = (Number.parseFloat(e.target.value) / 100) * duration
    audio.currentTime = newTime
    setCurrentTime(newTime)
  }

  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newVolume = Number.parseFloat(e.target.value) / 100
    setVolume(newVolume)
    if (audioRef.current) {
      audioRef.current.volume = newVolume
    }
  }

  const formatTime = (time: number) => {
    const minutes = Math.floor(time / 60)
    const seconds = Math.floor(time % 60)
    return `${minutes}:${seconds.toString().padStart(2, "0")}`
  }

  if (!audioUrl) {
    return (
      <div className={`border border-gray-600 rounded-lg p-4 bg-gray-900/30 backdrop-blur-sm ${className}`}>
        <div className="text-gray-400 text-center font-mono">{label ? `${label} - ` : ""}NO AUDIO LOADED</div>
      </div>
    )
  }

  return (
    <div className={`audio-player border border-cyan-400/50 rounded-lg p-4 bg-black/30 backdrop-blur-sm ${className}`}>
      <audio ref={audioRef} src={audioUrl} />

      {/* Label */}
      {label && (
        <div className="text-cyan-400 font-mono text-sm mb-3 flex items-center gap-2">
          <span className="w-2 h-2 bg-cyan-400 rounded-full animate-pulse"></span>
          {label}
        </div>
      )}

      {/* Controls */}
      <div className="flex items-center gap-4 mb-3">
        {/* Play/Pause button */}
        <button
          onClick={togglePlay}
          className="w-10 h-10 rounded-full bg-gradient-to-r from-pink-500 to-cyan-400 flex items-center justify-center text-black font-bold hover:scale-110 transition-transform"
        >
          {isPlaying ? "⏸" : "▶"}
        </button>

        {/* Time display */}
        <div className="text-white font-mono text-sm">
          {formatTime(currentTime)} / {formatTime(duration)}
        </div>

        {/* Volume control */}
        <div className="flex items-center gap-2 ml-auto">
          <span className="text-cyan-400 text-sm">🔊</span>
          <div className="relative w-20">
            <div className="w-full h-1 bg-gray-700 rounded-lg overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-cyan-400 to-pink-500 transition-all duration-100"
                style={{ width: `${volume * 100}%` }}
              ></div>
            </div>
            <input
              type="range"
              min="0"
              max="100"
              value={volume * 100}
              onChange={handleVolumeChange}
              className="absolute top-0 left-0 w-full h-1 bg-transparent appearance-none cursor-pointer"
            />
          </div>
        </div>
      </div>

      {/* Progress bar */}
      <div className="relative">
        <div className="w-full h-2 bg-gray-700 rounded-lg overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-pink-500 to-cyan-400 transition-all duration-100"
            style={{ width: `${duration ? (currentTime / duration) * 100 : 0}%` }}
          ></div>
        </div>
        <input
          type="range"
          min="0"
          max="100"
          value={duration ? (currentTime / duration) * 100 : 0}
          onChange={handleSeek}
          className="absolute top-0 left-0 w-full h-2 bg-transparent appearance-none cursor-pointer"
        />
      </div>

      {/* Waveform visualization placeholder */}
      <div className="mt-3 h-16 bg-black/50 rounded border border-cyan-400/30 flex items-center justify-center">
        <div className="flex items-end gap-1 h-8">
          {Array.from({ length: 32 }).map((_, i) => (
            <div
              key={i}
              className="w-1 bg-gradient-to-t from-cyan-400 to-pink-500 rounded-t animate-pulse"
              style={{
                height: `${Math.random() * 100}%`,
                animationDelay: `${i * 50}ms`,
              }}
            ></div>
          ))}
        </div>
      </div>

      {label === "STEGO AUDIO" && (
        <Button
          variant="secondary"
          className="mt-3 w-10 h-10 rounded-full bg-green-500 hover:bg-green-600 text-white p-0 flex items-center justify-center"
          onClick={onClick}
        >
          <ArrowDown size={16} />
        </Button>
      )}
    </div>
  )
}

export default AudioPlayer
