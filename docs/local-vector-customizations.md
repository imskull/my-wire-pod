# Local Vector customizations

This fork adds three features to the `chipper` module:

- `server.advertise_ip` pins the local IPv4 address used in the setup URL and mDNS announcements. It must be an address assigned to the WirePod host; an invalid address fails visibly instead of silently advertising a VPN interface.
- With Vosk, `en-US` selected, and Knowledge Graph enabled, the same audio is recognized by the English and Chinese models. A confidence-based choice keeps English built-in commands available while allowing Chinese questions to reach the LLM. Both Vosk models must already be installed. The active built-in intent table remains English; Chinese custom intents are separate.
- Chinese LLM sentences are streamed correctly, including Chinese punctuation or a final sentence without punctuation. On Windows, replies containing Han characters use a local Chinese voice, while English replies use Vector's voice.

For Windows Chinese speech, copy [`chipper/windows/chinese-tts.ps1`](../chipper/windows/chinese-tts.ps1) next to the running `chipper.exe`. The script currently selects the Windows `Microsoft Huihui Desktop` voice. Install or select an available Chinese Windows voice before using it on another PC. When building the separate Windows WirePod wrapper, point its `github.com/kercre123/wire-pod/chipper` module dependency at this fork's `chipper` directory (for example with a local `replace` directive), then copy the script alongside the built executable.

Configure the address in WirePod's `apiConfig.json` under `server`, using the host's actual LAN address, for example:

```json
"advertise_ip": "192.168.1.10"
```

If DHCP assigns a new address, update this value and restart WirePod. Keep API keys and the full configuration file out of Git.

For bilingual LLM replies, select `en-US` in WirePod and configure the LLM's prompt in the web UI. A suitable prompt is:

> You are Vector, a lively, curious little robot companion. Reply in the same language as the user: Chinese questions in Chinese, English questions in English. Answer directly in one or two short, natural sentences, or tell a slightly longer story when asked. Add a playful reaction or fitting robot animation when appropriate. Do not use emoji.

The language choice is heuristic, so short or unclear speech may be misclassified. Check the WirePod log's `Bilingual VOSK selection` and `Transcribed text` lines when tuning phrases. The Windows Chinese speech path was tested with the installed Huihui voice and a Vector 2; other voices and operating systems need separate verification.

The following Custom Intents were configured through the web UI on the original installation. They are user settings, not repository defaults, so add them only where desired:

| Utterances | Intent |
| --- | --- |
| `过来`, `到我这里来` | `intent_imperative_come` |
| `后退`, `退后` | `intent_imperative_backup` |
| `回去充电` | `intent_system_charger` |
| `跳支舞` | `intent_imperative_dance` |

English mode already has built-in English equivalents for these commands. Chinese Vosk may insert spaces between words, so an exact Chinese Custom Intent phrase may need an additional spaced variant.
