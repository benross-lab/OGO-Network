import math
from fractions import Fraction as F

r = F(9,10); N_H = 40; N_G = 110
HYB_A = [10000 + (60900 * r**n).__floor__() if n < N_H else 0 for n in range(120)]
GEO   = [(100000 * r**n).__floor__() if n < N_G else 0 for n in range(120)]

def run(R, D, F, fee_term, rule='RF', fee_offset=0):
    # v03 reconstruction: post-inflow balance, Trail_4 incl. current period,
    # float treasury share R_t*0.1, fee = terminal*(1-exp(-t/20)), treasury fee share 10%
    fees = [fee_term * (1 - math.exp(-(t+fee_offset)/20.0)) for t in range(120)]
    I = [R[t]/10.0 + 0.1*fees[t] for t in range(120)]
    B = 0.0; total = 0.0; bal = []
    for t in range(120):
        B = B + I[t]
        window = I[max(0,t-3):t+1]
        trail = sum(window)/len(window) if window else 0.0
        five = math.floor(0.05*B)
        room = max(0.0, B - F)
        if rule == 'RF':   spend = min(float(D), max(trail, five), room)
        elif rule == 'FIXED': spend = min(float(D), room)
        elif rule == 'TRAIL': spend = min(float(D), trail, room)
        elif rule == 'BAL5':  spend = min(float(D), five, room)
        elif rule == 'P4':    spend = min(float(D), trail, five, room)
        B = B - spend; total += spend; bal.append(B)
    return bal, total, sum(I)

for ft in (2462, 2462.5):
    print("=== fee terminal %s (Low), winning variant ===" % ft)
    for (R, name, D) in ((HYB_A,'HYB',1500),(HYB_A,'HYB',2500),(GEO,'GEO',1500)):
        bal, tot, tin = run(R, D, 10000, ft)
        print("  %s RF D=%d: total=%.4f  B39=%.2f  B109=%.2f  B119=%.4f  inflow=%.4f" % (
            name, D, tot, bal[39], bal[109], bal[119], tin))
    print("  DELTA (HYB 2500 - 1500) = %.6f" % (
        run(HYB_A,2500,10000,ft)[1] - run(HYB_A,1500,10000,ft)[1]))
    print("  targets: HYB1500 tot 114511 B39 46982 | HYB2500 tot 114511 B39 27023 | GEO1500 tot 114507")
    print("  section 6 (D=2500, Low, HYB):")
    for rule in ('FIXED','TRAIL','BAL5','P4'):
        bal, tot, tin = run(HYB_A, 2500, 10000, ft, rule)
        print("    %-6s total=%.2f B39=%.2f B119=%.2f (targets: FIXED 114511, TRAIL 98582/27538/25930, BAL5 114511/32544, P4 86202/39917/38309)" % (rule, tot, bal[39], bal[119]))
    print("  Mod/High HYB D=1500:")
    for ft2, nm, tb39, tb119 in ((12312,'MOD',None,44065),(36938,'HIGH',123760,289198)):
        bal, tot, tin = run(HYB_A, 1500, 10000, ft2)
        print("    %s total=%.2f B39=%.2f B119=%.2f (targets: tot 178500, B39 %s, B119 %s)" % (nm, tot, bal[39], bal[119], tb39, tb119))
