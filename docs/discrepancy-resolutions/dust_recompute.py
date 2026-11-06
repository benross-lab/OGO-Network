from fractions import Fraction as F

r = F(9,10)
N = 40

def sched(base, B0, n_periods=N):
    # R_n = base + floor(B0 * r^n), exact rational arithmetic
    return [base + (B0 * r**n).__floor__() if isinstance(B0, F) else base + int(B0 * r**n) for n in range(n_periods)]

def analytic_total(base, B0):
    # S_target = base*N + B0*(1-r^N)/(1-r)
    return base*N + B0*(1 - r**N)/(1 - r)

# Variant A: integer B0 = 60900 (prototype Go code, deep-dive, models_test.go)
B0_int = 60900
amounts_A = [10000 + (B0_int * r**n).__floor__() for n in range(N)]
tot_A = sum(amounts_A)
print("A integer B0=60900: R0=%d R39=%d total=%d analytic_target=%s dust_vs_1e6=%d" % (
    amounts_A[0], amounts_A[39], tot_A, analytic_total(10000, B0_int), 1000000 - tot_A))

# Variant B: exact rational B0 making S_target exactly 1,000,000
B0_rat = F(600000) / ((1 - r**N)/(1 - r))
print("B rational B0 =", float(B0_rat), " exact:", B0_rat)
amounts_B = [10000 + (B0_rat * r**n).__floor__() for n in range(N)]
tot_B = sum(amounts_B)
print("B rational B0: R0=%d R39=%d total=%d dust_vs_1e6=%d" % (amounts_B[0], amounts_B[39], tot_B, 1000000 - tot_B))

# Variant C: base 0%, integer B0 = 101500
amounts_C = [(101500 * r**n).__floor__() for n in range(N)]
tot_C = sum(amounts_C)
print("C base0 integer B0=101500: R0=%d R39=%d total=%d analytic_target=%s dust_vs_1e6=%d" % (
    amounts_C[0], amounts_C[39], tot_C, analytic_total(0, 101500), 1000000 - tot_C))

# Variant D: base 0%, exact rational B0 making S_target exactly 1,000,000
B0_rat0 = F(1000000) / ((1 - r**N)/(1 - r))
print("D rational B0(base0) =", float(B0_rat0))
amounts_D = [(B0_rat0 * r**n).__floor__() for n in range(N)]
tot_D = sum(amounts_D)
print("D rational B0 base0: R0=%d R39=%d total=%d dust_vs_1e6=%d" % (amounts_D[0], amounts_D[39], tot_D, 1000000 - tot_D))

# Geometric r=0.9 check: R0=100000
amounts_G = [(100000 * r**n).__floor__() for n in range(110)]
print("G geometric r=0.9 R0=100000 N=110: total=%d dust_vs_1e6=%d R109=%d" % (sum(amounts_G), 1000000-sum(amounts_G), amounts_G[109]))

# Treasury inflow check: sum floor(R_n/10) for variant A and B
print("treasury inflow sum floor(R/10): A=%d B=%d" % (sum(a//10 for a in amounts_A), sum(a//10 for a in amounts_B)))
print("treasury inflow sum floor(R/10): C(base0)=%d D(base0)=%d" % (sum(a//10 for a in amounts_C), sum(a//10 for a in amounts_D)))
print("geometric treasury inflow sum floor(R/10): %d" % sum(a//10 for a in amounts_G))
