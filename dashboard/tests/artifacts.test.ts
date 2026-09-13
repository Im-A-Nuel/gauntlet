import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, writeFile, rm } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { parseRun, score, compare, verdict } from '../lib/schema';
import { loadRuns } from '../lib/runs';

function fixture() {
  return {schemaVersion:1,runId:'100-abc',createdAt:'2026-09-13T15:00:00Z',baseRef:'main',headSha:'abc',trigger:'manual',threshold:80,
    changedFiles:['src/a.ts'],totals:{mutants:3,killed:1,timeout:1,survived:1,noCoverage:0,trustScore:66.7,lineCoverage:null,durationMs:100},
    files:[{path:'src/a.ts',trustScore:66.7,mutants:['killed','timeout','survived'].map((status,i)=>({id:String(i),mutator:'ConditionalExpression',line:1,status,original:'a > b',mutated:'a < b'}))}]};
}
test('timeouts count as detection; an empty denominator is unscored',()=>{
  assert.equal(score(24,2,34),43.3);assert.equal(score(0,0,0),null);
  assert.equal(verdict(parseRun(fixture())),'fail');
});
test('inconsistent measurements and unknown statuses are rejected',()=>{
  const bad=fixture();bad.totals.trustScore=100;assert.throws(()=>parseRun(bad));
  const badStatus=fixture();badStatus.files[0].mutants[0].status='pending';assert.throws(()=>parseRun(badStatus));
  const duplicate=fixture();duplicate.files[0].mutants[1].id='0';assert.throws(()=>parseRun(duplicate));
});
test('comparison distinguishes changed source scope and revisions',()=>{
  const a=parseRun(fixture()),b=parseRun(fixture());b.headSha='different';
  assert.equal(compare(a,b).comparable,false);assert.equal(compare(a,a).delta.trustScore,0);
});
test('artifact reader fails explicitly on corrupt live data and never replaces it with samples',async()=>{
  const dir=await mkdtemp(path.join(os.tmpdir(),'gauntlet-dashboard-'));
  try{
    assert.deepEqual((await loadRuns(dir)).runs,[]);
    await writeFile(path.join(dir,'100-abc.json'),JSON.stringify(fixture()));
    assert.equal((await loadRuns(dir)).source,'live');
    assert.equal((await loadRuns(dir)).runs[0].totals.trustScore,66.7);
    await writeFile(path.join(dir,'broken.json'),'{not json');
    await assert.rejects(loadRuns(dir),/Invalid run artifact/);
    await assert.rejects(loadRuns(path.join(dir,'missing')),/Cannot read/);
  }finally{await rm(dir,{recursive:true,force:true})}
});
