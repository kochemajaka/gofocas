#ifndef GOFOCAS_HELPERS_H
#define GOFOCAS_HELPERS_H

#include "fwlib32.h"

/* Helpers to access ODB* struct fields that are awkward from CGo due to
   anonymous unions / bitfields. */

/* ODBPOS — position per axis (cnc_rdposition returns array of ODBPOS) */
static long  focas_abs(ODBPOS *p)      { return p->abs.data; }
static long  focas_mach(ODBPOS *p)     { return p->mach.data; }
static long  focas_rel(ODBPOS *p)      { return p->rel.data; }
static long  focas_dist(ODBPOS *p)     { return p->dist.data; }
static short focas_abs_dec(ODBPOS *p)  { return p->abs.dec; }
static short focas_unit(ODBPOS *p)     { return p->abs.unit; }
static char  focas_name(ODBPOS *p)     { return p->abs.name; }
static char  focas_suf(ODBPOS *p)      { return p->abs.suff; }

/* ODBST — statinfo fields (default/else branch: hdck,tmmode,aut,run,...) */
static short focas_st_tmmode(ODBST *s)    { return s->tmmode; }
static short focas_st_autoact(ODBST *s)   { return s->aut; }
static short focas_st_runstate(ODBST *s)  { return s->run; }
static short focas_st_axis(ODBST *s)      { return s->motion; }
static short focas_st_edit(ODBST *s)      { return s->edit; }
static short focas_st_mstb(ODBST *s)      { return s->mstb; }
static short focas_st_emer(ODBST *s)      { return s->emergency; }
static short focas_st_alarm(ODBST *s)     { return s->alarm; }

/* ODBDGN / IODBPSD — union access (CGo cannot access C unions directly) */
static long focas_dgn_ldata(ODBDGN *d)    { return d->u.ldata; }
static long focas_psd_ldata(IODBPSD *p)   { return p->u.ldata; }

/* ODBSYS — sysinfo fields */
static short focas_sys_addinfo(ODBSYS *s) { return s->addinfo; }
static short focas_sys_maxaxis(ODBSYS *s) { return s->max_axis; }
static char  focas_sys_cnc0(ODBSYS *s)    { return s->cnc_type[0]; }
static char  focas_sys_cnc1(ODBSYS *s)    { return s->cnc_type[1]; }
static char  focas_sys_mt0(ODBSYS *s)     { return s->mt_type[0]; }
static char  focas_sys_mt1(ODBSYS *s)     { return s->mt_type[1]; }
static short focas_sys_axes(ODBSYS *s)    { return (short)(s->axes[0] - '0') * 10 + (short)(s->axes[1] - '0'); }

#endif /* GOFOCAS_HELPERS_H */
