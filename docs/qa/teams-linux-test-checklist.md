# QA Checklist — Teams Linux Remediation

## Test Plan

### Pre-requisites
- [ ] Chrome extension loaded in browser
- [ ] Native host (verbalizer) installed and running
- [ ] Daemon (verbalizerd) installed and running
- [ ] ffprobe available (for audio validation)
- [ ] Audio sources available on Linux (pactl or wpctl working)

### Test Scenarios

#### T1: Join call directly (no prejoin)
1. Open Teams Web (teams.microsoft.com or teams.live.com)
2. Join a meeting directly
3. Verify CALL_STARTED is logged with correlation ID
4. Wait for call to end
5. Verify CALL_ENDED is logged
6. Check recording file exists in recordings/
7. Check transcript generated in transcripts/

#### T2: Join via prejoin screen
1. Open Teams Web
2. Join a meeting that shows prejoin screen
3. Click "Join now" button
4. Verify CALL_STARTED fires after joining, not before
5. Verify call recording works

#### T3: Reconnect during call (network blip)
1. Start a call
2. Simulate network disconnect (or toggle WiFi)
3. Wait for Teams to reconnect
4. Verify no duplicate CALL_STARTED events
5. Verify recording continues (single file)

#### T4: Long call stability (>30 minutes)
1. Join a call
2. Wait 30+ minutes
3. Verify no false end detection
4. Verify recording is continuous

#### T5: Audio source validation
1. Run preflight check (or start recording)
2. Verify source discovery finds monitor sources
3. Verify recording has actual audio content (not silence)
4. Verify audio is from the call (not microphone)

### Expected Outcomes
- [ ] No duplicate CALL_STARTED for a single call
- [ ] No duplicate CALL_ENDED for a single call
- [ ] Recording file has audio content
- [ ] Transcript generated successfully
- [ ] Log shows correlation ID throughout

### Evidence Collection
For each test, capture:
1. Console log output (structured JSON logs)
2. Recording file path
3. Transcript file path
4. Any error messages