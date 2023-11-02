import pandas as pd
import matplotlib.pyplot as plt

# Load the CSV data into a pandas DataFrame
df = pd.read_csv('results.csv')

# Group by 'Party Setup' and calculate the mean for each metric
avg_times = df.groupby('Party Setup').mean()

# X-axis labels
x_labels = ['3-of-2', '4-of-2', '5-of-2', '6-of-2']

# Plotting
plt.figure(figsize=(10, 6))

# Extract Time plot
plt.plot(x_labels, avg_times['Extract Time (µs)'], color='red', marker='o', label='Extract Time')
# Lookup Time plot
plt.plot(x_labels, avg_times['Lookup Time (µs)'], color='blue', marker='o', label='Lookup Time')

# Calculate the overall averages for each function
avg_extract_time = avg_times['Extract Time (µs)'].mean()
avg_lookup_time = avg_times['Lookup Time (µs)'].mean()
avg_combined_time = avg_extract_time + avg_lookup_time

# Add the averages to the top of the graph
plt.figtext(0.15, 0.28, f"Extract Filter Avg: {avg_extract_time:.2f}µs", horizontalalignment='left', color='red')
plt.figtext(0.15, 0.25, f"Lookup Avg: {avg_lookup_time:.2f}µs", horizontalalignment='left', color='blue')
plt.figtext(0.15, 0.22, f"Combined Avg: {avg_combined_time:.2f}µs", horizontalalignment='left', color='green')

# Setting labels, title, and legend
plt.xlabel('Party Sets')
plt.ylabel('Time (µs)')
plt.title('Average Time vs Party Sets')
plt.legend()

# Save the plot as SVG
plt.tight_layout()
plt.savefig('graph.svg', format='svg')
plt.show()
