from fileinput import filename
import json
import sys
import essentia
essentia.logLevel = 'silent'
from essentia.standard import MonoLoader, KeyExtractor, RhythmExtractor2013
import numpy as np

if len(sys.argv) < 2:
        print("Usage: python extract.py <filename>", file=sys.stderr)
        sys.exit(1)

filename = sys.argv[1]
loader = MonoLoader(filename=filename)
audio = loader()

start_time = 10
duration = 20

sample_rate = 44100
start_sample = int(start_time * sample_rate)
end_sample = start_sample + int(duration * sample_rate)

audio_segment = audio[start_sample:end_sample]

key_extractor = KeyExtractor(profileType='edmm')
key, scale, strength = key_extractor(audio_segment)

rhythm_extractor = RhythmExtractor2013()
bpm, _, _, _, _ = rhythm_extractor(audio_segment)

result = {
    "key": key,
    "bpm": round(bpm)
}
print(json.dumps(result))